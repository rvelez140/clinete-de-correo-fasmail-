import 'dart:convert';
import 'package:fasmail_app/services/api_client.dart';
import 'package:fasmail_app/models/dashboard_stats.dart';
import 'package:fasmail_app/models/user.dart';
import 'package:fasmail_app/config/api_config.dart';

class AdminService {
  final ApiClient _client;

  AdminService({required ApiClient client}) : _client = client;

  Future<DashboardStats?> getDashboardStats() async {
    final response = await _client.get(ApiConfig.dashboardPath);
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body)['data'];
      return DashboardStats.fromJson(data);
    }
    return null;
  }

  Future<Map<String, String>?> getSettings() async {
    final response = await _client.get(ApiConfig.settingsPath);
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body)['data'];
      return Map<String, String>.from(data);
    }
    return null;
  }

  Future<bool> updateSetting(String key, String value) async {
    final response =
        await _client.put(ApiConfig.settingsPath, {'key': key, 'value': value});
    return response.statusCode == 200;
  }

  Future<List<User>?> getUsers({int page = 1, int limit = 20}) async {
    final response =
        await _client.get('${ApiConfig.usersPath}?page=$page&limit=$limit');
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body)['data'] as List;
      return data.map((u) => User.fromJson(u)).toList();
    }
    return null;
  }
}
