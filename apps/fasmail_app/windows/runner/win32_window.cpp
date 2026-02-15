#include "win32_window.h"
#include <dwmapi.h>
#include <flutter_windows.h>

namespace {
constexpr const wchar_t kWindowClassName[] = L"FLUTTER_RUNNER_WIN32_WINDOW";
using EnableNonClientDpiScaling = BOOL __stdcall(HWND hwnd);

static int Scale(int source, double scale_factor) {
  return static_cast<int>(source * scale_factor);
}

static void CenterAndShow(HWND hwnd) {
  HMONITOR monitor = MonitorFromWindow(hwnd, MONITOR_DEFAULTTONEAREST);
  MONITORINFO mi = {};
  mi.cbSize = sizeof(mi);
  if (GetMonitorInfo(monitor, &mi)) {
    RECT rc;
    GetWindowRect(hwnd, &rc);
    int w = rc.right - rc.left;
    int h = rc.bottom - rc.top;
    int x = (mi.rcWork.right - mi.rcWork.left - w) / 2 + mi.rcWork.left;
    int y = (mi.rcWork.bottom - mi.rcWork.top - h) / 2 + mi.rcWork.top;
    MoveWindow(hwnd, x, y, w, h, TRUE);
  }
  ShowWindow(hwnd, SW_SHOW);
  UpdateWindow(hwnd);
}
}  // namespace

Win32Window::Win32Window() {}
Win32Window::~Win32Window() { Destroy(); }

bool Win32Window::Create(const std::wstring& title,
                         const Point& origin, const Size& size) {
  Destroy();

  const wchar_t* window_class =
      WindowClassRegistrar::GetInstance()->GetWindowClass();

  double scale_factor = FlutterDesktopGetDpiForHWND(nullptr) / 96.0;

  HWND window = CreateWindow(
      window_class, title.c_str(), WS_OVERLAPPEDWINDOW,
      Scale(origin.x, scale_factor), Scale(origin.y, scale_factor),
      Scale(size.width, scale_factor), Scale(size.height, scale_factor),
      nullptr, nullptr, GetModuleHandle(nullptr), this);

  if (!window) return false;

  UpdateTheme(window);
  return OnCreate();
}

void Win32Window::Show() {
  CenterAndShow(window_handle_);
}

void Win32Window::SetQuitOnClose(bool quit) { quit_on_close_ = quit; }

HWND Win32Window::GetHandle() { return window_handle_; }

bool Win32Window::OnCreate() { return true; }

void Win32Window::OnDestroy() {}

RECT Win32Window::GetClientArea() {
  RECT frame;
  GetClientRect(window_handle_, &frame);
  return frame;
}

void Win32Window::SetChildContent(HWND content) {
  child_content_ = content;
  if (content != nullptr) {
    SetParent(content, window_handle_);
    RECT frame = GetClientArea();
    MoveWindow(content, frame.left, frame.top,
               frame.right - frame.left, frame.bottom - frame.top, true);
  }
}

LRESULT Win32Window::MessageHandler(HWND hwnd, UINT const message,
                                    WPARAM const wparam,
                                    LPARAM const lparam) noexcept {
  switch (message) {
    case WM_DESTROY:
      window_handle_ = nullptr;
      OnDestroy();
      if (quit_on_close_) {
        PostQuitMessage(0);
      }
      return 0;

    case WM_SIZE: {
      RECT frame = GetClientArea();
      if (child_content_ != nullptr) {
        MoveWindow(child_content_, frame.left, frame.top,
                   frame.right - frame.left, frame.bottom - frame.top, TRUE);
      }
      return 0;
    }

    case WM_ACTIVATE:
      if (child_content_ != nullptr) {
        SetFocus(child_content_);
      }
      return 0;

    case WM_DWMCOLORIZATIONCOLORCHANGED:
      UpdateTheme(hwnd);
      return 0;
  }

  return DefWindowProc(window_handle_, message, wparam, lparam);
}

void Win32Window::Destroy() {
  if (window_handle_) {
    DestroyWindow(window_handle_);
    window_handle_ = nullptr;
  }
}

void Win32Window::UpdateTheme(HWND hwnd) {
  BOOL use_dark = FALSE;
  DwmSetWindowAttribute(hwnd, 20 /*DWMWA_USE_IMMERSIVE_DARK_MODE*/,
                         &use_dark, sizeof(use_dark));
}
