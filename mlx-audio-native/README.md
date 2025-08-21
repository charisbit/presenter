# MLX-Audio Native macOS Service

这个目录包含在 macOS 上原生运行的 MLX-Audio TTS 服务。

## 前提条件

- macOS 13.5+ (推荐 macOS 14 Sonoma)
- Apple Silicon (M1/M2/M3/M4)
- Python 3.9+

## 安装步骤

1. 创建虚拟环境：
```bash
cd mlx-audio-native
python3 -m venv venv
source venv/bin/activate
```

2. 安装依赖：
```bash
pip install -r requirements.txt
```

3. 启动服务：
```bash
python app.py
```

**注意**：首次启动时会自动下载日语词典（UniDic），请耐心等待。

## 服务配置

- 端口：8881
- API 兼容 Docker 版本
- 使用 MLX 框架的全部性能优势

## API 端点

- `GET /health` - 健康检查
- `GET /models` - 列出可用模型
- `POST /api/tts` - 语音合成
- `GET /audio/{filename}` - 获取音频文件