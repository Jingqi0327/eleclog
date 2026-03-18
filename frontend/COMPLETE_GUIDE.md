# 前端 API 对接完整指南

## ✅ 完成的改进

### 1. **API 客户端集中管理** 
已创建 `src/client/` 文件夹，统一管理所有 API 请求：

```
src/client/
├── api.ts      # 配置了拦截器的 axios 实例
└── index.ts    # 导出接口
```

### 2. **自动环境切换**
- **开发环境** → `http://localhost:8080`（.env.development）
- **生产环境** → `/api`（.env.production） 

```bash
# 开发时
npm run dev
# 自动读取 .env.development，API 指向 http://localhost:8080

# 构建时
npm run build
# 自动读取 .env.production，API 指向 /api（由 nginx 代理）
```

### 3. **更新的组件列表**

| 组件 | 改进 |
|------|------|
| LoginUser.vue | 使用 apiClient，移除硬编码 URL |
| UserProfileMenu.vue | 简化认证头处理 |
| RoomAddForm.vue | 集中 API 调用 |
| RoomDeleteTable.vue | 使用 apiClient.delete() |
| RoomListView.vue | 简化实现 |
| RoomDetailView.vue | 统一 API 调用 |

### 4. **Nginx 代理优化**

```nginx
location /api/ {
    proxy_pass http://api:8080/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## 🚀 使用方式

### 开发环境

```bash
cd frontend
npm install
npm run dev
```

访问 http://localhost:5173，自动代理请求到 http://localhost:8080

### 生产环境 - Docker 构建

```bash
# 构建镜像
docker build -f frontend/Dockerfile -t elec_log_frontend:latest ./frontend

# 运行容器（需要与后端在同一 Docker 网络）
docker run --name frontend \
  -p 8080:8080 \
  --network=eleclog_default \
  elec_log_frontend:latest
```

构建时会自动使用 `.env.production`，API 指向 `/api`

## 📝 如何新增 API 调用

### ❌ 旧方式（已过时）
```typescript
// 不要这样写
import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

await axios.post(`${API_BASE}/users/login`, {
  username,
  password,
}, {
  headers: { Authorization: `Bearer ${token}` }
})
```

### ✅ 新方式（推荐）
```typescript
// 这样写，简洁且自动处理认证
import { apiClient } from '@/client'

await apiClient.post('/users/login', {
  username,
  password,
})
```

**优势：**
- Token 自动添加到所有请求
- 401 错误自动处理
- 开发/生产环境自动切换
- 无需重复定义 API_BASE

## 🔧 工作原理

### 请求拦截器
所有请求自动添加认证 token：
```typescript
// 自动执行
config.headers.Authorization = `Bearer ${token}`
```

### 响应拦截器
401 错误自动处理：
- 清除过期 token
- 触发 `auth-expired` 事件
- 可配置路由守卫进行重定向

### 环境变量
```env
# .env.development
VITE_API_BASE_URL=http://localhost:8080

# .env.production  
VITE_API_BASE_URL=/api
```

## 📦 Docker Compose 示例

```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_SOURCE=...

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    depends_on:
      - api
    networks:
      - default
```

nginx.conf 配置代理将 `/api/*` 请求转发给 `http://api:8080/`

## 🧪 验证

### 开发环境测试
```bash
npm run dev
# 打开浏览器 DevTools → Network
# 确认登录请求发送到 http://localhost:8080/users/login
```

### 生产环境测试
```bash
npm run build
# 检查 dist/ 中的 .js 文件
# 确认 API 地址是相对路径 /api/...
# 或使用 strings dist/assets/index-*.js | grep api
```

## 🐛 常见问题

### Q: 生产环境中 API 调用返回 404
**A:** 检查 nginx.conf 中 proxy_pass 的路径是否正确
```nginx
# ❌ 错误 - 少了末尾斜杠
proxy_pass http://api:8080;

# ✓ 正确
proxy_pass http://api:8080/;
```

### Q: 登录后返回 401
**A:** 确认 token 已保存到 localStorage
```javascript
// 在浏览器控制台检查
localStorage.getItem('accesstoken')
```

### Q: 开发环境无法连接到后端
**A:** 确认后端运行在 http://localhost:8080
```bash
curl http://localhost:8080/health
```

## 📚 文件清单

```
frontend/
├── .env.development         # 开发环境配置 ✓
├── .env.production          # 生产环境配置 ✓
├── src/
│   ├── client/
│   │   ├── api.ts           # API 客户端 ✓
│   │   └── index.ts         # 导出 ✓
│   ├── components/
│   │   ├── LoginUser.vue    # ✓ 更新
│   │   ├── UserProfileMenu.vue # ✓ 更新
│   │   └── rooms/
│   │       ├── RoomAddForm.vue     # ✓ 更新
│   │       └── RoomDeleteTable.vue # ✓ 更新
│   └── views/
│       ├── RoomListView.vue        # ✓ 更新
│       └── RoomDetailView.vue      # ✓ 更新
├── Dockerfile               # 已优化 ✓
├── nginx.conf               # 已优化 ✓
├── package.json             # 无需改动
└── API_SETUP.md             # 本文档
```

## 🎯 总结

✅ **已完成改进：**
- [x] 集中管理 axios 配置
- [x] 自动环境切换
- [x] Token 自动处理
- [x] 更新所有组件引用
- [x] 优化 nginx 代理
- [x] 创建环境配置文件
- [x] 构建测试通过

**下一步（可选）：**
- [ ] 添加 API 错误拦截和通用提示
- [ ] 实现请求重试机制
- [ ] 添加请求超时提醒
- [ ] 编写 API 层单元测试
