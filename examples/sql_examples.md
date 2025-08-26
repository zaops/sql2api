# SQL2API 使用示例

本文档提供了 SQL2API 系统的详细使用示例，包括所有 SQL 端点的使用方法。

## 认证

所有 SQL 端点都需要 API Key 认证。在请求头中包含您的 API Key：

```bash
X-API-Key: your-api-key-here
```

## 1. 通用 SQL 查询端点

### 端点
```
POST /api/v1/sql
```

### 原生 SQL 查询示例

#### 基本查询
```json
{
  "database_type": "postgres",
  "sql": "SELECT id, name, category FROM items WHERE active = $1",
  "params": {
    "active": true
  }
}
```

#### 带分页的查询
```json
{
  "database_type": "postgres",
  "sql": "SELECT * FROM items WHERE category = $1",
  "params": {
    "category": "electronics"
  },
  "pagination": {
    "page": 1,
    "page_size": 20
  },
  "sort": {
    "sort_by": "created_at",
    "sort_order": "desc"
  }
}
```

### 结构化查询示例

#### SELECT 查询
```json
{
  "database_type": "postgres",
  "query": {
    "table": "items",
    "action": "select",
    "fields": ["id", "name", "category", "created_at"],
    "where": {
      "active": true,
      "category": "electronics"
    },
    "order_by": [
      {"field": "created_at", "order": "desc"},
      {"field": "name", "order": "asc"}
    ],
    "limit": 50
  }
}
```

#### 聚合查询
```json
{
  "database_type": "postgres",
  "query": {
    "table": "items",
    "action": "select",
    "fields": ["category", "COUNT(*) as count"],
    "where": {
      "active": true
    },
    "group_by": ["category"],
    "having": {
      "count": "> 10"
    }
  }
}
```

### 响应示例
```json
{
  "success": true,
  "message": "Query executed successfully",
  "data": [
    {
      "id": 1,
      "name": "Laptop",
      "category": "electronics",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "name": "Phone",
      "category": "electronics",
      "created_at": "2024-01-14T15:20:00Z"
    }
  ],
  "columns": ["id", "name", "category", "created_at"],
  "total": 2,
  "page": 1,
  "page_size": 20,
  "execution_time": 15.5,
  "timestamp": "2024-01-15T12:00:00Z"
}
```

## 2. 批量 SQL 操作端点

### 端点
```
POST /api/v1/sql/batch
```

### 事务模式批量操作
```json
{
  "database_type": "postgres",
  "transactional": true,
  "operations": [
    {
      "database_type": "postgres",
      "sql": "INSERT INTO items (name, category) VALUES ($1, $2)",
      "params": {
        "name": "New Laptop",
        "category": "electronics"
      }
    },
    {
      "database_type": "postgres",
      "sql": "UPDATE items SET active = $1 WHERE id = $2",
      "params": {
        "active": true,
        "id": 1
      }
    },
    {
      "database_type": "postgres",
      "sql": "DELETE FROM items WHERE id = $1",
      "params": {
        "id": 999
      }
    }
  ]
}
```

### 非事务模式批量操作
```json
{
  "database_type": "postgres",
  "transactional": false,
  "continue_on_error": true,
  "operations": [
    {
      "database_type": "postgres",
      "query": {
        "table": "items",
        "action": "insert",
        "data": {
          "name": "Item 1",
          "category": "books"
        }
      }
    },
    {
      "database_type": "postgres",
      "query": {
        "table": "items",
        "action": "insert",
        "data": {
          "name": "Item 2",
          "category": "electronics"
        }
      }
    }
  ]
}
```

### 批量响应示例
```json
{
  "success": true,
  "message": "Batch executed successfully",
  "results": [
    {
      "index": 0,
      "success": true,
      "affected_rows": 1,
      "execution_time": 5.2
    },
    {
      "index": 1,
      "success": true,
      "affected_rows": 1,
      "execution_time": 3.8
    },
    {
      "index": 2,
      "success": false,
      "affected_rows": 0,
      "error": {
        "code": 4002,
        "message": "Record not found"
      },
      "execution_time": 2.1
    }
  ],
  "total_affected_rows": 2,
  "executed_count": 2,
  "failed_count": 1,
  "execution_time": 11.1,
  "timestamp": "2024-01-15T12:05:00Z"
}
```

## 3. 便捷插入端点

### 端点
```
POST /api/v1/sql/insert
```

### 基本插入
```json
{
  "database_type": "postgres",
  "table": "items",
  "data": {
    "name": "New Product",
    "category": "electronics",
    "description": "A great new product",
    "price": 299.99,
    "active": true
  }
}
```

### 带冲突处理的插入
```json
{
  "database_type": "postgres",
  "table": "items",
  "data": {
    "name": "Unique Product",
    "category": "electronics",
    "sku": "PROD-001"
  },
  "on_conflict": "ignore",
  "return_fields": ["id", "created_at"]
}
```

### 插入响应示例
```json
{
  "success": true,
  "message": "Insert executed successfully",
  "data": [
    {
      "id": 123,
      "created_at": "2024-01-15T12:10:00Z"
    }
  ],
  "affected_rows": 1,
  "execution_time": 8.3,
  "timestamp": "2024-01-15T12:10:00Z"
}
```

## 4. 批量插入端点

### 端点
```
POST /api/v1/sql/batch-insert
```

### 批量插入示例
```json
{
  "database_type": "postgres",
  "table": "items",
  "data": [
    {
      "name": "Product 1",
      "category": "electronics",
      "price": 199.99
    },
    {
      "name": "Product 2",
      "category": "books",
      "price": 29.99
    },
    {
      "name": "Product 3",
      "category": "electronics",
      "price": 399.99
    }
  ],
  "on_conflict": "update",
  "return_fields": ["id", "name"]
}
```

### 批量插入响应示例
```json
{
  "success": true,
  "message": "Batch insert executed successfully",
  "data": [
    {"id": 124, "name": "Product 1"},
    {"id": 125, "name": "Product 2"},
    {"id": 126, "name": "Product 3"}
  ],
  "affected_rows": 3,
  "execution_time": 12.7,
  "timestamp": "2024-01-15T12:15:00Z"
}
```

## 错误响应示例

### 语法错误 (4001)
```json
{
  "success": false,
  "error": {
    "code": 4001,
    "message": "SQL syntax error",
    "details": "syntax error at or near \"SELEC\"",
    "sql_state": "42601"
  },
  "timestamp": "2024-01-15T12:20:00Z"
}
```

### 权限错误 (4003)
```json
{
  "success": false,
  "error": {
    "code": 4003,
    "message": "Permission denied",
    "details": "Access to table 'users' is not allowed"
  },
  "timestamp": "2024-01-15T12:25:00Z"
}
```

### 超时错误 (4006)
```json
{
  "success": false,
  "error": {
    "code": 4006,
    "message": "Query timeout",
    "details": "Query execution exceeded 30 seconds limit"
  },
  "timestamp": "2024-01-15T12:30:00Z"
}
```

## Oracle 数据库示例

### Oracle 查询示例
```json
{
  "database_type": "oracle",
  "sql": "SELECT id, name, category FROM items WHERE active = :1",
  "params": {
    "active": 1
  },
  "pagination": {
    "page": 1,
    "page_size": 10
  }
}
```

### Oracle 插入示例
```json
{
  "database_type": "oracle",
  "table": "items",
  "data": {
    "name": "Oracle Product",
    "category": "database",
    "active": 1
  },
  "return_fields": ["id", "created_date"]
}
```

## 权限配置

确保您的 API Key 具有相应的权限：

- `sql.query`: 查询权限
- `sql.insert`: 插入权限
- `sql.update`: 更新权限
- `sql.delete`: 删除权限
- `sql.batch`: 批量操作权限
- `sql.*`: 所有 SQL 权限

## 最佳实践

1. **使用参数化查询**：始终使用 `params` 字段传递参数，避免 SQL 注入
2. **合理设置分页**：对于大结果集，使用分页避免内存问题
3. **使用事务**：对于相关的多个操作，使用事务确保数据一致性
4. **错误处理**：根据错误码实现适当的错误处理逻辑
5. **性能监控**：关注 `execution_time` 字段，优化慢查询
