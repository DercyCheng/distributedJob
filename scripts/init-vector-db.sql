-- 初始化向量数据库
-- 启用pgvector扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 创建向量存储表
CREATE TABLE IF NOT EXISTS vector_store (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    metadata JSONB,
    embedding vector(1536), -- 默认使用OpenAI的1536维向量
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建向量索引
CREATE INDEX IF NOT EXISTS vector_store_embedding_idx ON vector_store USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- 创建命名空间表，用于组织不同的向量集合
CREATE TABLE IF NOT EXISTS vector_namespace (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建向量与命名空间的关联表
CREATE TABLE IF NOT EXISTS vector_namespace_mapping (
    id SERIAL PRIMARY KEY,
    vector_id INTEGER REFERENCES vector_store(id) ON DELETE CASCADE,
    namespace_id INTEGER REFERENCES vector_namespace(id) ON DELETE CASCADE,
    UNIQUE(vector_id, namespace_id)
);

-- 创建索引用于加速查询
CREATE INDEX IF NOT EXISTS idx_vector_namespace_mapping_vector_id ON vector_namespace_mapping(vector_id);
CREATE INDEX IF NOT EXISTS idx_vector_namespace_mapping_namespace_id ON vector_namespace_mapping(namespace_id);

-- 创建文档表，用于存储原始文档
CREATE TABLE IF NOT EXISTS document_store (
    id SERIAL PRIMARY KEY,
    title TEXT,
    content TEXT NOT NULL,
    metadata JSONB,
    source TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建文档与向量的关联表
CREATE TABLE IF NOT EXISTS document_vector_mapping (
    id SERIAL PRIMARY KEY,
    document_id INTEGER REFERENCES document_store(id) ON DELETE CASCADE,
    vector_id INTEGER REFERENCES vector_store(id) ON DELETE CASCADE,
    UNIQUE(document_id, vector_id)
);

-- 创建索引用于加速查询
CREATE INDEX IF NOT EXISTS idx_document_vector_mapping_document_id ON document_vector_mapping(document_id);
CREATE INDEX IF NOT EXISTS idx_document_vector_mapping_vector_id ON document_vector_mapping(vector_id);

-- 创建一个函数来更新更新时间
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为需要自动更新更新时间的表创建触发器
CREATE TRIGGER update_vector_store_updated_at
BEFORE UPDATE ON vector_store
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();

CREATE TRIGGER update_document_store_updated_at
BEFORE UPDATE ON document_store
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();

-- 创建一个默认的命名空间
INSERT INTO vector_namespace (name, description)
VALUES ('default', 'Default namespace for vector embeddings')
ON CONFLICT (name) DO NOTHING;

-- 创建函数来执行向量相似度搜索
CREATE OR REPLACE FUNCTION search_vectors(
    query_embedding vector(1536),
    namespace_name TEXT,
    limit_count INTEGER DEFAULT 10,
    similarity_threshold FLOAT DEFAULT 0.7
)
RETURNS TABLE (
    id INTEGER,
    content TEXT,
    metadata JSONB,
    similarity FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        vs.id,
        vs.content,
        vs.metadata,
        1 - (vs.embedding <=> query_embedding) AS similarity
    FROM
        vector_store vs
    JOIN
        vector_namespace_mapping vnm ON vs.id = vnm.vector_id
    JOIN
        vector_namespace vn ON vnm.namespace_id = vn.id
    WHERE
        vn.name = namespace_name
        AND 1 - (vs.embedding <=> query_embedding) > similarity_threshold
    ORDER BY
        similarity DESC
    LIMIT
        limit_count;
END;
$$ LANGUAGE plpgsql;
