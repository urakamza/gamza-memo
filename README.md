# 🥔 감자 메모

> Windows용 가볍고 심플한 스티커 메모 앱

## 📋 기능

- **서식 지원** — 굵게, 기울임, 밑줄, 취소선, 글머리 기호, 번호 목록
- **이미지 삽입** — 붙여넣기, 드래그 앤 드롭, 파일 선택으로 이미지 삽입
- **이미지 뷰어** — 이미지 더블클릭으로 별도 창에서 크게 보기 (확대/축소, 드래그)
- **색상 변경** — 7가지 프리셋 색상 + 커스텀 색상 지원
- **항상 위** — 다른 창 위에 고정
- **글꼴 설정** — 시스템 폰트 목록에서 글꼴, 굵기, 크기, 자간, 행간 설정
- **다크/라이트 테마**
- **시스템 트레이** — 트레이에서 새 메모 추가, 메모 목록 열기
- **시작 시 자동 실행** — Windows 시작 시 자동으로 실행 가능

## 💾 설치 방법

1. [Releases](../../releases) 페이지에서 최신 버전의 `gamzamemo.exe`를 다운로드합니다.
2. 원하는 폴더에 저장 후 실행합니다.
3. 별도 설치 과정 없이 바로 사용할 수 있습니다.

**시스템 요구사항**
- Windows 10 이상
- WebView2 런타임 (Windows 11은 기본 포함, Windows 10은 [여기서 설치](https://developer.microsoft.com/ko-kr/microsoft-edge/webview2/))

## 🗂️ 메모 백업 방법

메모 데이터는 아래 경로에 저장됩니다:

```
%APPDATA%\gamzamemo\
(C:\Users\사용자명\AppData\Roaming\gamzamemo\)
  notes.db       ← 메모 내용 및 이미지
  config.json    ← 설정 파일
```

**백업**
1. `Win + R` 또는 `시작버튼 우클릭-실행` → `%APPDATA%\uraksticky` 입력 후 Enter
2. `notes.db` 파일을 안전한 곳에 복사

**복원**
1. 복원할 PC에서 감자 메모를 실행한 후 완전히 종료합니다. (트레이 아이콘 우클릭 → 종료)
2. `%APPDATA%\uraksticky\notes.db`를 백업 파일로 교체합니다.
3. 감자 메모를 다시 실행합니다

> ⚠️ 복원 시 현재 메모가 모두 백업 시점으로 되돌아갑니다.

## 📦 사용한 도구

- [Wails v3](https://v3.wails.io/) — Go + WebView2 기반 데스크탑 앱 프레임워크
- [Svelte](https://svelte.dev/) — 프론트엔드
- [Tiptap](https://tiptap.dev/) — 리치 텍스트 에디터
- [SQLite](https://www.sqlite.org/) — 로컬 데이터 저장
- DirectWrite — 시스템 폰트 목록 조회 (Windows API)

## 📄 라이선스

MIT License © 2026 [urakamza](https://urakamza.kr)

```
이 소프트웨어는 MIT 라이선스 하에 자유롭게 사용, 수정, 배포할 수 있습니다.
```

---

[홈페이지](https://urakamza.kr)
