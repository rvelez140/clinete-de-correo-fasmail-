import 'package:flutter/material.dart';
import 'package:fasmail_app/services/auth_service.dart';
import 'package:fasmail_app/models/user.dart';

class AuthProvider extends ChangeNotifier {
  final AuthService _authService = AuthService();
  bool _isLoading = false;
  String? _error;

  AuthService get authService => _authService;
  User? get currentUser => _authService.currentUser;
  bool get isLoggedIn => _authService.isLoggedIn;
  bool get isLoading => _isLoading;
  String? get error => _error;
  String get serverUrl => _authService.serverUrl;

  Future<bool> checkSavedAuth() async {
    _isLoading = true;
    notifyListeners();

    final result = await _authService.checkSavedAuth();

    _isLoading = false;
    notifyListeners();
    return result;
  }

  Future<bool> login(String email, String password, {String? serverUrl}) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    final success =
        await _authService.login(email, password, serverUrl: serverUrl);

    if (!success) {
      _error = 'Credenciales inválidas';
    }

    _isLoading = false;
    notifyListeners();
    return success;
  }

  Future<void> logout() async {
    await _authService.logout();
    notifyListeners();
  }

  Future<void> setServerUrl(String url) async {
    await _authService.setServerUrl(url);
    notifyListeners();
  }

  Future<Map<String, dynamic>?> validateSetupToken(
      String token, String server) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    final result = await _authService.validateSetupToken(token, server);
    if (result == null) {
      _error = 'Token inválido o expirado';
    }

    _isLoading = false;
    notifyListeners();
    return result;
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }
}
