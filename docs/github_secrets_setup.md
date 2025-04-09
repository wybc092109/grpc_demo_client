# GitHub Secrets 配置指南

为了确保 CI/CD 流程能够正常运行，需要在 GitHub 仓库中配置以下 Secrets：

## 必要的 Secrets 配置

1. **TENCENT_NAMESPACE**

   - 说明：腾讯云容器镜像服务的命名空间
   - 获取方式：在腾讯云容器镜像服务控制台查看或创建命名空间

2. **TENCENT_CLOUD_USERNAME**

   - 说明：腾讯云容器镜像服务的用户名
   - 获取方式：在腾讯云访问管理控制台获取

3. **TENCENT_CLOUD_PASSWORD**

   - 说明：腾讯云容器镜像服务的密码
   - 获取方式：在腾讯云访问管理控制台获取

4. **TENCENT_CLOUD_API_TOKEN**

   - 说明：用于访问腾讯云 API 的令牌
   - 获取方式：在腾讯云访问管理控制台创建 API 密钥

5. **TENCENT_TKE_WEBHOOK_URL**
   - 说明：TKE 集群的 Webhook URL，用于触发集群更新
   - 获取方式：在 TKE 集群配置中获取或创建 Webhook

## 配置步骤

1. 打开 GitHub 仓库页面
2. 点击 Settings 标签
3. 在左侧菜单中选择 Secrets and variables > Actions
4. 点击 New repository secret 按钮
5. 依次添加上述所有 Secrets，确保名称完全匹配

## 注意事项

- 所有的 Secret 值都应该是敏感信息，请妥善保管
- 确保所有 Secret 都已正确配置，否则 CI/CD 流程可能会失败
- 定期检查并更新过期的密钥和令牌
