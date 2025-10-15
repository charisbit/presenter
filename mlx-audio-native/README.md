# MLX-Audio Native macOS Service

这个目录包含在 macOS 上原生运行的 MLX-Audio TTS 服务，使用 Kokoro-82M 模型提供高质量的文本转语音功能。

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

**注意**：首次启动时会自动下载日语词典（UniDic）和中文分词词典，请耐心等待。

## 服务配置

- 端口：8881
- API 兼容 Docker 版本
- 使用 MLX 框架的全部性能优势
- 支持日语和中文文本处理

## 支持的语音

### 日语语音 (lang_code='j')
- **女性**: `jf_alpha`, `jf_gongitsune`, `jf_nezumi`, `jf_tebukuro`
- **男性**: `jm_kumo`

### 中文语音 (lang_code='z')
- **女性**: `zf_xiaobei`, `zf_xiaoni`, `zf_xiaoxiao`, `zf_xiaoyi`
- **男性**: `zm_yunjian`, `zm_yunxi`, `zm_yunxia`, `zm_yunyang`

### 英语语音 (lang_code='a')
- **女性**: `af_heart`, `af_alloy`, `af_aoede`, `af_bella`, `af_jessica`, `af_kore`, `af_nicole`, `af_nova`, `af_river`, `af_sarah`, `af_sky`
- **男性**: `am_adam`, `am_echo`, `am_eric`, `am_fenrir`, `am_liam`, `am_michael`, `am_onyx`, `am_puck`, `am_santa`

## API 端点

- `GET /health` - 健康检查
- `GET /models` - 列出可用模型
- `POST /api/tts` - 语音合成
- `GET /audio/{filename}` - 获取音频文件

### TTS 请求示例

```bash
# 日语合成
curl -X POST "http://localhost:8881/api/tts" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "こんにちは、これはテストです。",
    "language": "ja", 
    "voice": "female",
    "speed": 1.0
  }'

# 中文合成
curl -X POST "http://localhost:8881/api/tts" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "你好，这是一个测试。",
    "language": "zh", 
    "voice": "female",
    "speed": 1.0
  }'
```

## 技术特性

- 使用 MLX-Audio 框架，针对 Apple Silicon 优化
- 支持 Kokoro-82M 模型的高质量语音合成
- 自动缓存生成的音频文件
- 支持多种音频格式（WAV、MP3等）
- 完整的日语和中文文本处理支持
