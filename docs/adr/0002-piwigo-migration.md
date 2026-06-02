# ADR 0002: Piwigo 迁移功能

## Status

Accepted

## Context

用户希望将 Piwigo 照片库迁移到 Flickr。Piwigo 是一个自托管的开源照片管理系统。

经过调研，Piwigo 提供完整的 REST API（`ws.php`），支持：
- `pwg.categories.getList` — 获取所有相册及层级
- `pwg.categories.getImages` — 获取相册中的照片
- `pwg.images.getInfo` — 获取照片详细元数据
- `pwg.tags.getList` — 获取所有标签
- `pwg.images.exist` — 通过 md5sum 检查照片是否已存在
- `pwg.session.login` — 认证

## Decision

使用 Piwigo REST API 实现 Piwigo→Flickr 迁移功能，不再直接连接 MySQL。

Piwigo 已有官方 Flickr2Piwigo 插件（`Piwigo/Flickr2Piwigo`）支持 Flickr→Piwigo 方向，因此 flickr-cli 只实现 Piwigo→Flickr 单向迁移。

迁移流程：
1. 通过 `pwg.session.login` 认证
2. 通过 `pwg.categories.getList` 获取相册层级
3. 通过 `pwg.categories.getImages` 遍历每个相册的照片
4. 通过 `pwg.images.getInfo` 获取每张照片的元数据
5. 通过 `pwg.images.exist` 检查 md5sum 实现去重
6. 下载照片文件（通过 API 返回的 URL）
7. 上传到 Flickr 并创建对应相册

命令设计：
```bash
flickr piwigo import --url https://photos.example.com --user admin --password secret
flickr piwigo import --url ... --user ... --password ... --album "Vacation"
flickr piwigo import --url ... --user ... --password ... --album-prefix "Piwigo/"
```

密码通过 flag 传递（一次性迁移，不持久化 Piwigo 凭证）。

## Consequences

### Positive

- 不需要 MySQL 驱动依赖
- 不需要用户开放数据库端口
- 只需要 Piwigo URL + 用户名密码
- 复用 Piwigo 已有的查询能力
- 更安全（通过 HTTP 认证，而非数据库直连）

### Negative

- 需要 Piwigo 实例可访问
- 大量照片时 API 调用次数多（但支持分页）
