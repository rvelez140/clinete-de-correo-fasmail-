cmake_minimum_required(VERSION 3.14)
project(runner LANGUAGES CXX)

add_executable(${BINARY_NAME} WIN32
  "flutter_window.cpp"
  "main.cpp"
  "utils.cpp"
  "win32_window.cpp"
  "${FLUTTER_MANAGED_DIR}/generated_plugin_registrant.cc"
  "Runner.rc"
  "runner.exe.manifest"
)

target_compile_features(${BINARY_NAME} PRIVATE cxx_std_17)

target_link_libraries(${BINARY_NAME} PRIVATE flutter flutter_wrapper_plugin)

target_include_directories(${BINARY_NAME} PRIVATE
  "${CMAKE_SOURCE_DIR}"
)

if(CMAKE_BUILD_TYPE STREQUAL "Debug")
  target_compile_definitions(${BINARY_NAME} PRIVATE "DEBUG")
endif()
