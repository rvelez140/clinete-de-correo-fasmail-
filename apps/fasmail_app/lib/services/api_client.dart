import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:fasmail_app/services/auth_service.dart';
import 'package:fasmail_app/config/api_config.dart';

class ApiClient {
  final AuthService _authService;

  ApiClient({required AuthService authService}) : _authService = authService;

  String get _baseUrl => ApiConfig.apiUrl(_authService.serverUrl);

  Map<String, String> _headers(String? token) => {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        if (token != null) 'Authorization': 'Bearer $token',
      };

  Future<http.Response> get(String path) async {
    final token = await _authService.getAccessToken();
    var response = await http
        .get(Uri.parse('$_baseUrl$path'), headers: _headers(token))
        .timeout(ApiConfig.connectTimeout);

    if (response.statusCode == 401) {
      final refreshed = await _authService.refreshToken();
      if (refreshed) {
        final newToken = await _authService.getAccessToken();
        response = await http
            .get(Uri.parse('$_baseUrl$path'), headers: _headers(newToken))
            .timeout(ApiConfig.connectTimeout);
      }
    }

    return response;
  }

  Future<http.Response> post(String path, Map<String, dynamic> body) async {
    final token = await _authService.getAccessToken();
    var response = await http
        .post(Uri.parse('$_baseUrl$path'),
            headers: _headers(token), body: jsonEncode(body))
        .timeout(ApiConfig.connectTimeout);

    if (response.statusCode == 401) {
      final refreshed = await _authService.refreshToken();
      if (refreshed) {
        final newToken = await _authService.getAccessToken();
        response = await http
            .post(Uri.parse('$_baseUrl$path'),
                headers: _headers(newToken), body: jsonEncode(body))
            .timeout(ApiConfig.connectTimeout);
      }
    }

    return response;
  }

  Future<http.Response> put(String path, Map<String, dynamic> body) async {
    final token = await _authService.getAccessToken();
    var response = await http
        .put(Uri.parse('$_baseUrl$path'),
            headers: _headers(token), body: jsonEncode(body))
        .timeout(ApiConfig.connectTimeout);

    if (response.statusCode == 401) {
      final refreshed = await _authService.refreshToken();
      if (refreshed) {
        final newToken = await _authService.getAccessToken();
        response = await http
            .put(Uri.parse('$_baseUrl$path'),
                headers: _headers(newToken), body: jsonEncode(body))
            .timeout(ApiConfig.connectTimeout);
      }
    }

    return response;
  }

  Future<http.Response> delete(String path) async {
    final token = await _authService.getAccessToken();
    var response = await http
        .delete(Uri.parse('$_baseUrl$path'), headers: _headers(token))
        .timeout(ApiConfig.connectTimeout);

    if (response.statusCode == 401) {
      final refreshed = await _authService.refreshToken();
      if (refreshed) {
        final newToken = await _authService.getAccessToken();
        response = await http
            .delete(Uri.parse('$_baseUrl$path'), headers: _headers(newToken))
            .timeout(ApiConfig.connectTimeout);
      }
    }

    return response;
  }

  Future<http.Response> publicPost(
      String serverUrl, String path, Map<String, dynamic> body) async {
    final url = '${ApiConfig.apiUrl(serverUrl)}$path';
    return http
        .post(Uri.parse(url),
            headers: _headers(null), body: jsonEncode(body))
        .timeout(ApiConfig.connectTimeout);
  }

  Future<http.Response> publicGet(String serverUrl, String path) async {
    final url = '${ApiConfig.apiUrl(serverUrl)}$path';
    return http
        .get(Uri.parse(url), headers: _headers(null))
        .timeout(ApiConfig.connectTimeout);
  }
}
