import 'dart:convert';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'package:fasmail_app/config/api_config.dart';
import 'package:fasmail_app/models/user.dart';

class AuthService {
  static const _accessTokenKey = 'access_token';
  static const _refreshTokenKey = 'refresh_token';
  static const _serverUrlKey = 'server_url';
  static const _userDataKey = 'user_data';

  final FlutterSecureStorage _secureStorage = const FlutterSecureStorage();

  String _serverUrl = ApiConfig.defaultServerUrl;
  String? _accessToken;
  String? _refreshToken;
  User? _currentUser;

  String get serverUrl => _serverUrl;
  User? get currentUser => _currentUser;
  bool get isLoggedIn => _accessToken != null;

  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _serverUrl = prefs.getString(_serverUrlKey) ?? ApiConfig.defaultServerUrl;
    _accessToken = await _secureStorage.read(key: _accessTokenKey);
    _refreshToken = await _secureStorage.read(key: _refreshTokenKey);

    final userData = prefs.getString(_userDataKey);
    if (userData != null) {
      _currentUser = User.fromJson(jsonDecode(userData));
    }
  }

  Future<String?> getAccessToken() async {
    return _accessToken;
  }

  Future<void> setServerUrl(String url) async {
    _serverUrl = url;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_serverUrlKey, url);
  }

  Future<bool> login(String email, String password, {String? serverUrl}) async {
    if (serverUrl != null) {
      await setServerUrl(serverUrl);
    }

    final url = '${ApiConfig.apiUrl(_serverUrl)}${ApiConfig.loginPath}';
    final response = await http.post(
      Uri.parse(url),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: jsonEncode({'email': email, 'password': password}),
    );

    if (response.statusCode != 200) {
      return false;
    }

    final data = jsonDecode(response.body)['data'];
    _accessToken = data['access_token'];
    _refreshToken = data['refresh_token'];
    _currentUser = User.fromJson(data['user']);

    await _secureStorage.write(key: _accessTokenKey, value: _accessToken);
    await _secureStorage.write(key: _refreshTokenKey, value: _refreshToken);

    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_userDataKey, jsonEncode(data['user']));

    return true;
  }

  Future<bool> refreshToken() async {
    if (_refreshToken == null) return false;

    final url = '${ApiConfig.apiUrl(_serverUrl)}${ApiConfig.refreshPath}';
    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: jsonEncode({'refresh_token': _refreshToken}),
      );

      if (response.statusCode != 200) {
        await logout();
        return false;
      }

      final data = jsonDecode(response.body)['data'];
      _accessToken = data['access_token'];
      _refreshToken = data['refresh_token'];

      await _secureStorage.write(key: _accessTokenKey, value: _accessToken);
      await _secureStorage.write(key: _refreshTokenKey, value: _refreshToken);

      return true;
    } catch (e) {
      return false;
    }
  }

  Future<void> logout() async {
    if (_accessToken != null) {
      try {
        final url = '${ApiConfig.apiUrl(_serverUrl)}${ApiConfig.logoutPath}';
        await http.post(
          Uri.parse(url),
          headers: {
            'Authorization': 'Bearer $_accessToken',
            'Accept': 'application/json',
          },
        );
      } catch (_) {}
    }

    _accessToken = null;
    _refreshToken = null;
    _currentUser = null;

    await _secureStorage.delete(key: _accessTokenKey);
    await _secureStorage.delete(key: _refreshTokenKey);

    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_userDataKey);
  }

  Future<bool> checkSavedAuth() async {
    await init();
    if (_accessToken == null) return false;

    try {
      final url = '${ApiConfig.apiUrl(_serverUrl)}${ApiConfig.profilePath}';
      final response = await http.get(
        Uri.parse(url),
        headers: {
          'Authorization': 'Bearer $_accessToken',
          'Accept': 'application/json',
        },
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body)['data'];
        _currentUser = User.fromJson(data);
        return true;
      }

      return await refreshToken();
    } catch (e) {
      return false;
    }
  }

  Future<Map<String, dynamic>?> validateSetupToken(
      String token, String server) async {
    final url =
        '${ApiConfig.apiUrl(server)}${ApiConfig.validateTokenPath}?token=$token';
    try {
      final response = await http.get(
        Uri.parse(url),
        headers: {'Accept': 'application/json'},
      );

      if (response.statusCode == 200) {
        await setServerUrl(server);
        return jsonDecode(response.body)['data'];
      }
    } catch (e) {
      print('Error validating setup token: $e');
    }
    return null;
  }
}
