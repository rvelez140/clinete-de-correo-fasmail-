import 'dart:convert';
import 'package:fasmail_app/services/api_client.dart';
import 'package:fasmail_app/models/company.dart';
import 'package:fasmail_app/config/api_config.dart';

class CompanyService {
  final ApiClient _client;

  CompanyService({required ApiClient client}) : _client = client;

  Future<List<Company>?> getCompanies({int page = 1, int limit = 20}) async {
    final response = await _client
        .get('${ApiConfig.companiesPath}?page=$page&limit=$limit');
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body)['data'] as List;
      return data.map((c) => Company.fromJson(c)).toList();
    }
    return null;
  }

  Future<Company?> getCompany(String id) async {
    final response = await _client.get('${ApiConfig.companiesPath}/$id');
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body)['data'];
      return Company.fromJson(data);
    }
    return null;
  }

  Future<Company?> createCompany(Company company) async {
    final response =
        await _client.post(ApiConfig.companiesPath, company.toJson());
    if (response.statusCode == 201) {
      final data = jsonDecode(response.body)['data'];
      return Company.fromJson(data);
    }
    return null;
  }

  Future<bool> updateCompany(String id, Company company) async {
    final response =
        await _client.put('${ApiConfig.companiesPath}/$id', company.toJson());
    return response.statusCode == 200;
  }

  Future<bool> deleteCompany(String id) async {
    final response =
        await _client.delete('${ApiConfig.companiesPath}/$id');
    return response.statusCode == 200;
  }
}
