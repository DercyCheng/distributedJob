package vectorstore

import (
	"fmt"

	"distributedJob/internal/rag/embedding"
)

// New 创建一个新的向量存储
func New(config map[string]interface{}, embeddingProvider embedding.Provider) (VectorStore, error) {
	storeType, ok := config["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid vector store type")
	}

	switch storeType {
	case "memory":
		return NewMemoryVectorStore(), nil

	case "postgres":
		dimension := embeddingProvider.GetDimension()
		// 提取Postgres配置
		dbConfig := PostgresConfig{
			Dimension: dimension,
		}

		if connStr, ok := config["connection_string"].(string); ok {
			dbConfig.ConnectionString = connStr
		} else {
			return nil, fmt.Errorf("postgres vector store requires connection_string")
		}

		if tableName, ok := config["table_name"].(string); ok {
			dbConfig.TableName = tableName
		}

		if idCol, ok := config["id_column"].(string); ok {
			dbConfig.IDColumn = idCol
		}

		if vecCol, ok := config["vector_column"].(string); ok {
			dbConfig.VectorColumn = vecCol
		}

		if dataCol, ok := config["data_column"].(string); ok {
			dbConfig.DataColumn = dataCol
		}

		if metaCol, ok := config["metadata_column"].(string); ok {
			dbConfig.MetadataColumn = metaCol
		}

		return NewPostgresVectorStore(dbConfig)

	case "qdrant":
		// 实现Qdrant向量存储的创建逻辑
		return nil, fmt.Errorf("qdrant vector store not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported vector store type: %s", storeType)
	}
}
