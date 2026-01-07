# GitHub MCP Server with GHES Asynchronous Coding Agent

[![Go Report Card](https://goreportcard.com/badge/github.com/github/github-mcp-server)](https://goreportcard.com/report/github.com/github/github-mcp-server)

GitHub MCP Server는 AI 도구를 GitHub 플랫폼에 직접 연결하는 MCP(Model Context Protocol) 서버입니다. AI 에이전트, 어시스턴트, 챗봇이 자연어 상호작용을 통해 리포지토리와 코드 파일 읽기, 이슈 및 PR 관리, 코드 분석, 워크플로우 자동화를 수행할 수 있습니다.

## 🎯 주요 기능

### GHES 온프레미스 비동기 코딩 에이전트 지원 (feature/ghes-agent)

이 브랜치는 **GitHub Enterprise Server(GHES) 온프레미스 환경**에서도 비동기 코딩 에이전트를 사용할 수 있도록 확장된 기능을 제공합니다.

#### 새로운 기능

- **`assign_github_copilot_to_issue`**: GitHub 이슈를 자동으로 분석하고 코딩 에이전트(SWE Agent)에게 할당하여 비동기로 작업을 처리합니다.
- **SWE Agent 통합**: [HakjunMIN/swe-agent](https://github.com/HakjunMIN/swe-agent) 프로젝트를 REST API 서버로 구성하여 사용합니다.
- **Azure OpenAI 지원**: Azure OpenAI 서비스를 백엔드 LLM으로 사용하여 GHES 환경에서도 강력한 코딩 에이전트 기능을 제공합니다.
- **자동 PR 생성**: 코딩 에이전트가 이슈를 해결한 후 자동으로 Pull Request를 생성합니다.

### 기본 기능

- **리포지토리 관리**: 코드 탐색, 파일 검색, 커밋 분석, 프로젝트 구조 이해
- **이슈 & PR 자동화**: 이슈 및 풀 리퀘스트 생성, 업데이트, 관리
- **CI/CD & 워크플로우**: GitHub Actions 모니터링, 빌드 실패 분석, 릴리스 관리
- **코드 분석**: 보안 취약점 검토, Dependabot 알림, 코드 패턴 분석
- **팀 협업**: 토론 접근, 알림 관리, 팀 활동 분석

---

## 🏗️ 빌드 방법

### 사전 요구사항

- Go 1.24 이상
- Docker (선택 사항, 컨테이너 실행 시)

### 소스 코드 빌드

```bash
# 저장소 클론
git clone https://github.com/github/github-mcp-server.git
cd github-mcp-server

# feature/ghes-agent 브랜치 체크아웃
git checkout feature/ghes-agent

# 의존성 다운로드
go mod download

# 빌드
go build -v ./cmd/github-mcp-server

# 실행
./github-mcp-server stdio
```

### Docker 이미지 빌드

```bash
# Docker 이미지 빌드
docker build -t github-mcp-server .

# Docker 컨테이너 실행
docker run -i --rm \
  -e GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here \
  github-mcp-server
```

### 개발 및 테스트

```bash
# 코드 포맷 및 린트
script/lint

# 테스트 실행
script/test

# 문서 생성 (MCP 도구 변경 시)
script/generate-docs
```

---

## ⚙️ SWE Agent 구성

GHES 환경에서 비동기 코딩 에이전트를 사용하려면 [HakjunMIN/swe-agent](https://github.com/HakjunMIN/swe-agent) 프로젝트를 REST API 서버로 실행해야 합니다.

### SWE Agent REST API 서버 설정

1. **SWE Agent 저장소 클론**
   ```bash
   git clone https://github.com/HakjunMIN/swe-agent.git
   cd swe-agent
   ```

2. **환경 변수 설정** (`.env` 파일 생성)
   ```env
   AZURE_OPENAI_API_BASE=https://your-azure-openai-endpoint
   AZURE_OPENAI_API_KEY=your-azure-openai-key
   AZURE_OPENAI_MODEL=azure/gpt-4o
   AZURE_OPENAI_API_VERSION=2024-02-15-preview
   GITHUB_TOKEN=your-github-token
   ```

3. **REST API 서버 실행**
   ```bash
   # Python 가상 환경 생성
   python -m venv venv
   source venv/bin/activate

   # 의존성 설치
   pip install -r requirements.txt

   # API 서버 실행 (기본 포트: 8000)
   python -m swe_agent.api.server
   ```

4. **서버 확인**
   ```bash
   curl http://localhost:8000/health
   ```

---

## 🔧 MCP 설정 (mcp.json)

### VS Code 설정

VS Code의 MCP 설정은 `mcp.json` 파일에 저장됩니다. 일반적으로 다음 위치에 있습니다:
- **macOS/Linux**: `~/.config/Code/User/mcp.json`
- **Windows**: `%APPDATA%\Code\User\mcp.json`

### 기본 설정 (로컬 stdio)

```json
{
  "servers": {
    "github": {
      "command": "/Users/andy/works/ai/github-mcp-server/github-mcp-server",
      "args": ["stdio"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "your_github_pat_here"
      }
    }
  }
}
```

### GHES 환경 설정 (온프레미스)

```json
{
  "servers": {
    "github-ghes": {
      "command": "/Users/andy/works/ai/github-mcp-server/github-mcp-server",
      "args": [
        "stdio",
        "--gh-host=https://your-ghes-instance.com"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "your_ghes_pat_here",
        "GITHUB_HOST": "https://your-ghes-instance.com"
      }
    }
  }
}
```

### SWE Agent 통합 설정 (GHES + 코딩 에이전트)

```json
{
  "servers": {
    "github-ghes-agent": {
      "command": "/Users/andy/works/ai/github-mcp-server/github-mcp-server",
      "args": [
        "stdio",
        "--gh-host=https://your-ghes-instance.com"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "your_ghes_pat_here",
        "GITHUB_HOST": "https://your-ghes-instance.com",
        "SWE_AGENT_ENDPOINT": "http://localhost:8000",
        "AZURE_OPENAI_API_BASE": "https://your-azure-openai-endpoint",
        "AZURE_OPENAI_API_KEY": "your-azure-openai-key",
        "AZURE_OPENAI_MODEL": "azure/gpt-4o",
        "AZURE_OPENAI_API_VERSION": "2024-02-15-preview",
        "GITHUB_TOKEN": "your_ghes_pat_here"
      }
    }
  }
}
```

### Docker 사용 설정

```json
{
  "servers": {
    "github-docker": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "GITHUB_PERSONAL_ACCESS_TOKEN"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "your_github_pat_here"
      }
    }
  }
}
```

### 환경 변수 설명

| 환경 변수 | 설명 | 필수 여부 |
|---------|-----|---------|
| `GITHUB_PERSONAL_ACCESS_TOKEN` | GitHub Personal Access Token | ✅ 필수 |
| `GITHUB_HOST` | GHES 호스트 URL (예: `https://github.company.com`) | GHES 사용 시 필수 |
| `SWE_AGENT_ENDPOINT` | SWE Agent REST API 엔드포인트 (기본값: `http://localhost:8000`) | 코딩 에이전트 사용 시 필수 |
| `AZURE_OPENAI_API_BASE` | Azure OpenAI API 베이스 URL | SWE Agent 사용 시 필수 |
| `AZURE_OPENAI_API_KEY` | Azure OpenAI API 키 | SWE Agent 사용 시 필수 |
| `AZURE_OPENAI_MODEL` | Azure OpenAI 모델 이름 (예: `azure/gpt-4o`) | SWE Agent 사용 시 필수 |
| `AZURE_OPENAI_API_VERSION` | Azure OpenAI API 버전 | SWE Agent 사용 시 필수 |
| `GITHUB_TOKEN` | SWE Agent가 사용할 GitHub 토큰 | SWE Agent 사용 시 필수 |
| `GITHUB_TOOLSETS` | 활성화할 툴셋 (쉼표로 구분) | 선택 사항 |
| `GITHUB_READ_ONLY` | 읽기 전용 모드 활성화 (`1` 설정 시) | 선택 사항 |

---

## 🚀 사용 방법

### 1. 기본 사용

MCP 서버를 설정한 후, VS Code의 Copilot Chat에서 Agent 모드를 활성화하고 GitHub 관련 작업을 요청할 수 있습니다.

```
@github 이 리포지토리의 최근 이슈를 보여줘
@github main.go 파일의 내용을 읽어줘
@github 새로운 브랜치를 만들고 파일을 수정해줘
```

### 2. 비동기 코딩 에이전트 사용

SWE Agent를 구성한 후, 이슈를 코딩 에이전트에게 할당할 수 있습니다:

```
@github issue #123을 코딩 에이전트에게 할당해줘
```

코딩 에이전트는 다음 작업을 자동으로 수행합니다:
1. 이슈 내용 분석
2. 관련 코드 탐색 및 이해
3. 버그 수정 또는 기능 구현
4. 테스트 작성 (필요 시)
5. Pull Request 자동 생성

### 3. 진행 상황 확인

```
@github SWE Agent 작업 상태를 확인해줘
@github 최근 생성된 PR을 보여줘
```

---

## 📋 MCP 도구 목록

### 코딩 에이전트 도구

- **`assign_github_copilot_to_issue`**: GitHub 이슈를 SWE Agent에게 할당하여 자동으로 해결
- **`assign_copilot_to_issue`**: (레거시) 코딩 에이전트에게 이슈 할당

### 리포지토리 관리

- `get_file_contents`: 파일 내용 조회
- `create_or_update_file`: 파일 생성 또는 업데이트
- `delete_file`: 파일 삭제
- `push_files`: 여러 파일을 한 번에 커밋
- `search_code`: 코드 검색
- `get_repository_tree`: 리포지토리 트리 조회

### 이슈 & PR 관리

- `issue_read`: 이슈 조회
- `issue_write`: 이슈 생성/업데이트
- `pull_request_read`: PR 조회
- `create_pull_request`: PR 생성
- `merge_pull_request`: PR 병합
- `pull_request_review_write`: PR 리뷰 작성

### GitHub Actions

- `list_workflows`: 워크플로우 목록 조회
- `list_workflow_runs`: 워크플로우 실행 목록
- `get_workflow_run`: 워크플로우 실행 상세 정보
- `run_workflow`: 워크플로우 실행
- `rerun_workflow_run`: 워크플로우 재실행

전체 도구 목록은 [README.old.md](README.old.md)를 참조하세요.

---

## 🔐 보안 고려사항

### GitHub Personal Access Token 관리

1. **최소 권한 원칙**: 필요한 스코프만 부여
   - `repo`: 리포지토리 작업
   - `read:packages`: Docker 이미지 접근
   - `read:org`: 조직 팀 접근

2. **토큰 분리**: 프로젝트/환경별로 다른 PAT 사용

3. **정기적인 갱신**: 주기적으로 토큰 업데이트

4. **버전 관리 제외**: `.gitignore`에 설정 파일 추가
   ```bash
   echo ".env" >> .gitignore
   echo "mcp.json" >> .gitignore
   ```

5. **파일 권한 설정**
   ```bash
   chmod 600 ~/.config/Code/User/mcp.json
   ```

### SWE Agent 보안

- **네트워크 격리**: SWE Agent API 서버를 내부 네트워크에만 노출
- **인증 추가**: 프로덕션 환경에서는 API 키 또는 OAuth 인증 구성
- **로그 모니터링**: SWE Agent 작업 로그를 정기적으로 검토

---

## 🧪 테스트

```bash
# 단위 테스트 실행
go test ./...

# 특정 패키지 테스트
go test ./pkg/github -v

# 레이스 컨디션 체크
script/test

# E2E 테스트 (GitHub PAT 필요)
GITHUB_MCP_SERVER_E2E_TOKEN=your_token go test -v --tags e2e ./e2e
```

---

## 🤝 기여하기

기여는 언제나 환영합니다! 자세한 내용은 [CONTRIBUTING.md](CONTRIBUTING.md)를 참조하세요.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests and linting (`script/test && script/lint`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

---

## 📄 라이선스

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 감사의 말

- [Model Context Protocol](https://modelcontextprotocol.io/) - MCP 프로토콜
- [go-github](https://github.com/google/go-github) - GitHub API 클라이언트
- [HakjunMIN/swe-agent](https://github.com/HakjunMIN/swe-agent) - SWE Agent REST API 구현

---

## 📞 지원

- [Issues](https://github.com/github/github-mcp-server/issues) - 버그 리포트 및 기능 요청
- [Discussions](https://github.com/github/github-mcp-server/discussions) - 질문 및 토론
- [SUPPORT.md](SUPPORT.md) - 지원 리소스
- [SECURITY.md](SECURITY.md) - 보안 정책

---

**참고**: 원본 README.md는 [README.old.md](README.old.md)에 백업되어 있습니다.
