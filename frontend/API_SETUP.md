# 前端 API 架构改进

## 概述
整合前端 axios 请求到统一的 API 客户端，支持开发和生产环境的自动切换。

## 目录结构
```
src/
├── client/                 # 新增 API 客户端文件夹
│   ├── api.ts             # 配置好的 axios 实例和拦截器
│   └── index.ts           # 导出接口
├── components/         
│   ├── LoginUser.vue      # 已更新 ✓
│   ├── UserProfileMenu.vue # 已更新 ✓
│   └── rooms/
│       ├── RoomAddForm.vue      # 已更新 ✓
│       └── RoomDeleteTable.vue  # 已更新 ✓
└── views/
    ├── RoomListView.vue        # 已更新 ✓
    └── RoomDetailView.vue      # 已更新 ✓
```

## 核心特性

### 1. 统一的 API 客户端 (`src/client/api.ts`)
- **基础 URL 自动切换**
  - 开发环境：`http://localhost:8080`（来自 .env.development）
  - 生产环境：`/api`（来自 .env.production，由 nginx 代理）
  - 手动配置优先：`VITE_API_BASE_URL` 环境变量

- **请求拦截器**
  - 自动添加认证 token 到请求头：`Authorization: Bearer <token>`
  - 从 localStorage 获取存储的 token

- **响应拦截器**
  - 自动处理 401 错误（token 过期）
  - 清除过期 token 并触发 `auth-expired` 事件

### 2. 环境变量配置

**.env.development** (开发环境)
```env
VITE_API_BASE_URL=http://localhost:8080
```

**.env.production** (生产环境 - nginx 构建)
```env
VITE_API_BASE_URL=/api
```

### 3. Nginx 代理配置

```nginx
location /api/ {
    proxy_pass http://api:8080/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## 使用方式

### 旧方式（已废弃）
```typescript
import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const response = await axios.post(`${API_BASE}/users/login`, {...})
```

### 新方式（推荐）
```typescript
import { apiClient } from '@/client'

const response = await apiClient.post('/users/login', {...})
// Token 自动添加，无需手动处理 headers
```

## 开发与构建

### 开发环境
```bash
npm run dev
# 自动使用 .env.development
# API 调用：http://localhost:8080/users/login
```

### 生产环境（nginx Docker 镜像）
```bash
npm run build
# 自动使用 .env.production
# 构建时 VITE_API_BASE_URL=/api 被编译到代码中
# API 调用：/api/users/login → nginx 代理到 http://api:8080/users/login
```

### Docker 构建命令
```bash
docker build -f frontend/Dockerfile -t elec_log_frontend:latest ./frontend
docker run --name frontend -p 8080:8080 --network=eleclog_default elec_log_frontend:latest
```

## 改进说明

| 问题 | 旧方式 | 新方式 |
|------|--------|--------|
| API 地址 | 各组件重复定义 | 统一在 `client/api.ts` |
| Token 处理 | 每个请求手动添加 header | 请求拦截器自动处理 |
| 401 处理 | 各组件分别处理 | 响应拦截器统一处理 |
| 环境切换 | 手动改代码 | 自动通过 .env 文件 |
| 代码重复 | 高 | 低 |

## 验证清单

- [x] 创建 `src/client/` 文件夹
- [x] 实现 `api.ts` 配置
- [x] 添加请求/响应拦截器
- [x] 创建 `.env.development` 和 `.env.production`
- [x] 更新所有组件使用 `apiClient`
- [x] 优化 nginx 代理配置
- [x] 删除各组件的 `API_BASE` 定义

## 下一步

1. ✅ 测试开发环境运行
2. ✅ 测试生产环境 nginx 构建
3. ✅ 验证登录和 token 管理
4. ✅ 检查所有 API 路由是否正常
