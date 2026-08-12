package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// 최소한 이 정도 픽셀은 화면에 보여야 "화면 안"으로 인정
const minVisibleMargin = 40

// clampToVisibleScreen은 저장된 창 좌표(x, y)가 현재 연결된 모니터 중
// 어디에도 충분히 걸치지 않으면, 안전한 기본 위치로 보정해서 돌려준다.
// 두 번째 반환값은 보정이 실제로 일어났는지 여부.
func clampToVisibleScreen(app *application.App, x, y, width, height int) (int, int, bool) {
	screens := app.Screen.GetAll()
	if len(screens) == 0 {
		// 화면 정보를 못 가져오면 원래 좌표 그대로 사용
		return x, y, false
	}

	winRect := application.Rect{X: x, Y: y, Width: width, Height: height}

	for _, screen := range screens {
		if screen == nil {
			continue
		}
		overlap := winRect.Intersect(screen.Bounds)
		if overlap.IsEmpty() {
			continue
		}
		// 타이틀바를 잡을 수 있을 정도로 충분히 겹치면 "보인다"고 판단
		if overlap.Width >= minVisibleMargin && overlap.Height >= minVisibleMargin {
			return x, y, false
		}
	}

	// 어떤 모니터와도 충분히 겹치지 않음 → 주 모니터 기준으로 보정
	primary := app.Screen.GetPrimary()
	if primary == nil {
		// 주 모니터 정보도 없으면 그냥 첫 번째 화면 기준
		primary = screens[0]
	}

	safeX := primary.Bounds.X + 100
	safeY := primary.Bounds.Y + 100
	return safeX, safeY, true
}
