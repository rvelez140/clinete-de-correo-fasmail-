class ApiConfig {
  static const String defaultServerUrl = 'http://localhost:8080';
  static const String apiBasePath = '/api/v1';
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);

  static String apiUrl(String serverUrl) => '$serverUrl$apiBasePath';

  // Auth endpoints
  static const String loginPath = '/auth/login';
  static const String refreshPath = '/auth/refresh';
  static const String logoutPath = '/auth/logout';
  static const String profilePath = '/auth/profile';

  // Admin endpoints
  static const String dashboardPath = '/admin/dashboard';
  static const String settingsPath = '/admin/settings';
  static const String usersPath = '/admin/users';
  static const String companiesPath = '/admin/companies';

  // Downloads endpoints
  static const String generateTokenPath = '/downloads/generate-token';
  static const String validateTokenPath = '/downloads/validate-token';

  // Email endpoints
  static const String emailAccountsPath = '/email/accounts';
}
