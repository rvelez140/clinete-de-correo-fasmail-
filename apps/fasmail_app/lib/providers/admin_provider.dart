import 'package:flutter/material.dart';
import 'package:fasmail_app/services/api_client.dart';
import 'package:fasmail_app/services/admin_service.dart';
import 'package:fasmail_app/services/company_service.dart';
import 'package:fasmail_app/services/auth_service.dart';
import 'package:fasmail_app/models/dashboard_stats.dart';
import 'package:fasmail_app/models/company.dart';
import 'package:fasmail_app/models/user.dart';

class AdminProvider extends ChangeNotifier {
  AdminService? _adminService;
  CompanyService? _companyService;

  DashboardStats? _stats;
  List<Company>? _companies;
  List<User>? _users;
  Map<String, String>? _settings;
  bool _isLoading = false;
  String? _error;

  DashboardStats? get stats => _stats;
  List<Company>? get companies => _companies;
  List<User>? get users => _users;
  Map<String, String>? get settings => _settings;
  bool get isLoading => _isLoading;
  String? get error => _error;

  void init(AuthService authService) {
    final client = ApiClient(authService: authService);
    _adminService = AdminService(client: client);
    _companyService = CompanyService(client: client);
  }

  Future<void> loadDashboard() async {
    if (_adminService == null) return;
    _isLoading = true;
    _error = null;
    notifyListeners();

    _stats = await _adminService!.getDashboardStats();
    if (_stats == null) {
      _error = 'Error al cargar estadísticas';
    }

    _isLoading = false;
    notifyListeners();
  }

  Future<void> loadCompanies() async {
    if (_companyService == null) return;
    _isLoading = true;
    notifyListeners();

    _companies = await _companyService!.getCompanies();

    _isLoading = false;
    notifyListeners();
  }

  Future<void> loadUsers() async {
    if (_adminService == null) return;
    _isLoading = true;
    notifyListeners();

    _users = await _adminService!.getUsers();

    _isLoading = false;
    notifyListeners();
  }

  Future<void> loadSettings() async {
    if (_adminService == null) return;
    _isLoading = true;
    notifyListeners();

    _settings = await _adminService!.getSettings();

    _isLoading = false;
    notifyListeners();
  }

  Future<bool> createCompany(Company company) async {
    if (_companyService == null) return false;
    final result = await _companyService!.createCompany(company);
    if (result != null) {
      await loadCompanies();
      return true;
    }
    return false;
  }

  Future<bool> deleteCompany(String id) async {
    if (_companyService == null) return false;
    final success = await _companyService!.deleteCompany(id);
    if (success) {
      await loadCompanies();
    }
    return success;
  }
}
