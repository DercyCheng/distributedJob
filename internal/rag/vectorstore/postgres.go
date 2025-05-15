package vectorstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgresVectorStore 实现基于PostgreSQL的向量存储
// 使用PostgreSQL的pgvector扩展
type PostgresVectorStore struct {
	db         *sql.DB
	tableName  string
	dimension  int
	idColumn   string
	vecColumn  string
	dataColumn string
	metaColumn string
}

// PostgresConfig PostgreSQL向量存储配置
type PostgresConfig struct {
	ConnectionString string
	TableName        string
	Dimension        int
	IDColumn         string
	VectorColumn     string
	DataColumn       string
	MetadataColumn   string
}

// NewPostgresVectorStore 创建新的PostgreSQL向量存储
func NewPostgresVectorStore(config PostgresConfig) (*PostgresVectorStore, error) {
	// 连接数据库
	db, err := sql.Open("postgres", config.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// 检查连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// 设置默认值
	tableName := config.TableName
	if tableName == "" {
		tableName = "vector_store"
	}

	dimension := config.Dimension
	if dimension <= 0 {
		dimension = 1536 // 默认维度
	}

	idColumn := config.IDColumn
	if idColumn == "" {
		idColumn = "id"
	}

	vecColumn := config.VectorColumn
	if vecColumn == "" {
		vecColumn = "embedding"
	}

	dataColumn := config.DataColumn
	if dataColumn == "" {
		dataColumn = "content"
	}

	metaColumn := config.MetadataColumn
	if metaColumn == "" {
		metaColumn = "metadata"
	}

	store := &PostgresVectorStore{
		db:         db,
		tableName:  tableName,
		dimension:  dimension,
		idColumn:   idColumn,
		vecColumn:  vecColumn,
		dataColumn: dataColumn,
		metaColumn: metaColumn,
	}

	// 初始化表
	if err := store.initTable(); err != nil {
		return nil, err
	}

	return store, nil
}

// initTable 初始化向量表
func (s *PostgresVectorStore) initTable() error {
	// 检查pgvector扩展是否可用
	_, err := s.db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("failed to create pgvector extension: %w", err)
	}

	// 创建表
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			%s TEXT PRIMARY KEY,
			%s TEXT NOT NULL,
			%s JSONB,
			%s vector(%d)
		)
	`, s.tableName, s.idColumn, s.dataColumn, s.metaColumn, s.vecColumn, s.dimension)

	_, err = s.db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create vector table: %w", err)
	}

	// 创建向量索引
	createIndexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_%s_idx ON %s USING hnsw (%s vector_cosine_ops)
	`, s.tableName, s.vecColumn, s.tableName, s.vecColumn)

	_, err = s.db.Exec(createIndexQuery)
	if err != nil {
		return fmt.Errorf("failed to create vector index: %w", err)
	}

	return nil
}

// Add 添加文档到向量存储
func (s *PostgresVectorStore) Add(ctx context.Context, documents []Document) error {
	if len(documents) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 准备插入语句
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (%s, %s, %s, %s) 
		VALUES ($1, $2, $3, $4) 
		ON CONFLICT (%s) DO UPDATE SET
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s,
			%s = EXCLUDED.%s
	`, s.tableName, s.idColumn, s.dataColumn, s.metaColumn, s.vecColumn,
		s.idColumn, s.dataColumn, s.dataColumn, s.metaColumn, s.metaColumn, s.vecColumn, s.vecColumn)

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// 插入每个文档
	for _, doc := range documents {
		// 确保文档有ID
		id := doc.ID
		if id == "" {
			id = uuid.New().String()
		}

		// 将元数据序列化为JSON
		metadataJSON, err := json.Marshal(doc.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		// 确保向量维度正确
		if len(doc.Vector) != s.dimension {
			return fmt.Errorf("vector dimension mismatch: expected %d, got %d", s.dimension, len(doc.Vector))
		}

		// 构造向量语句
		vecStr := fmt.Sprintf("[%s]", strings.Join(floatToStrings(doc.Vector), ","))

		// 执行插入
		_, err = stmt.ExecContext(ctx, id, doc.Content, metadataJSON, vecStr)
		if err != nil {
			return fmt.Errorf("failed to insert document: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Search 基于查询向量搜索最相似的文档
func (s *PostgresVectorStore) Search(ctx context.Context, queryVector []float32, limit int, filters map[string]interface{}) ([]SearchResult, error) {
	if len(queryVector) != s.dimension {
		return nil, fmt.Errorf("query vector dimension mismatch: expected %d, got %d", s.dimension, len(queryVector))
	}

	if limit <= 0 {
		limit = 10 // 默认限制
	}

	// 构造查询向量
	vecStr := fmt.Sprintf("[%s]", strings.Join(floatToStrings(queryVector), ","))

	// 构建查询
	query := fmt.Sprintf(`
		SELECT %s, %s, %s, 1 - (%s <=> $1::vector) as similarity
		FROM %s
	`, s.idColumn, s.dataColumn, s.metaColumn, s.vecColumn, s.tableName)

	// 添加过滤条件
	args := []interface{}{vecStr}
	argIndex := 2
	if filters != nil && len(filters) > 0 {
		whereConditions := []string{}
		for key, value := range filters {
			whereConditions = append(whereConditions, fmt.Sprintf("%s @> $%d::jsonb", s.metaColumn, argIndex))
			filterJSON, err := json.Marshal(map[string]interface{}{key: value})
			if err != nil {
				return nil, fmt.Errorf("failed to marshal filter: %w", err)
			}
			args = append(args, string(filterJSON))
			argIndex++
		}
		if len(whereConditions) > 0 {
			query += " WHERE " + strings.Join(whereConditions, " AND ")
		}
	}

	// 添加排序和限制
	query += fmt.Sprintf(" ORDER BY similarity DESC LIMIT %d", limit)

	// 执行查询
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	// 处理结果
	var results []SearchResult
	for rows.Next() {
		var id, content string
		var metadataBytes []byte
		var similarity float32

		if err := rows.Scan(&id, &content, &metadataBytes, &similarity); err != nil {
			return nil, fmt.Errorf("failed to scan result row: %w", err)
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		results = append(results, SearchResult{
			Document: Document{
				ID:       id,
				Content:  content,
				Metadata: metadata,
				// 向量通常不返回，因为很大
			},
			Score:    similarity,
			Distance: 1 - similarity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during result rows iteration: %w", err)
	}

	return results, nil
}

// Delete 删除指定ID的文档
func (s *PostgresVectorStore) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// 构造删除查询
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ANY($1)", s.tableName, s.idColumn)

	// 执行删除
	_, err := s.db.ExecContext(ctx, query, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	return nil
}

// Get 获取指定ID的文档
func (s *PostgresVectorStore) Get(ctx context.Context, id string) (Document, error) {
	query := fmt.Sprintf("SELECT %s, %s, %s FROM %s WHERE %s = $1",
		s.idColumn, s.dataColumn, s.metaColumn, s.tableName, s.idColumn)

	var docID, content string
	var metadataBytes []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(&docID, &content, &metadataBytes)
	if err != nil {
		if err == sql.ErrNoRows {
			return Document{}, fmt.Errorf("document with ID %s not found", id)
		}
		return Document{}, fmt.Errorf("failed to get document: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return Document{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return Document{
		ID:       docID,
		Content:  content,
		Metadata: metadata,
	}, nil
}

// List 列出所有文档（可分页）
func (s *PostgresVectorStore) List(ctx context.Context, offset, limit int) ([]Document, error) {
	if limit <= 0 {
		limit = 100 // 默认限制
	}

	query := fmt.Sprintf("SELECT %s, %s, %s FROM %s ORDER BY %s LIMIT $1 OFFSET $2",
		s.idColumn, s.dataColumn, s.metaColumn, s.tableName, s.idColumn)

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var id, content string
		var metadataBytes []byte

		if err := rows.Scan(&id, &content, &metadataBytes); err != nil {
			return nil, fmt.Errorf("failed to scan document row: %w", err)
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		documents = append(documents, Document{
			ID:       id,
			Content:  content,
			Metadata: metadata,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during document rows iteration: %w", err)
	}

	return documents, nil
}

// Count 获取文档总数
func (s *PostgresVectorStore) Count(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", s.tableName)

	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// Close 关闭数据库连接
func (s *PostgresVectorStore) Close() error {
	return s.db.Close()
}

// floatToStrings 将float32切片转换为字符串切片
func floatToStrings(floats []float32) []string {
	strs := make([]string, len(floats))
	for i, f := range floats {
		strs[i] = fmt.Sprintf("%f", f)
	}
	return strs
}

// DeleteCollection 删除整个集合
func (s *PostgresVectorStore) DeleteCollection(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", s.tableName))
	return err
}

// CreateCollection 创建一个新集合
func (s *PostgresVectorStore) CreateCollection(ctx context.Context, dimension int) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		%s TEXT PRIMARY KEY,
		%s TEXT,
		%s JSONB,
		%s vector(%d)
	)`, s.tableName, s.idColumn, s.dataColumn, s.metaColumn, s.vecColumn, dimension)

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// CollectionExists 检查集合是否存在
func (s *PostgresVectorStore) CollectionExists(ctx context.Context) (bool, error) {
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = $1
	)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, s.tableName).Scan(&exists)
	return exists, err
}
