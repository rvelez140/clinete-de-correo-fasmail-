import 'dart:convert';
import 'package:fasmail_app/services/api_client.dart';
import 'package:fasmail_app/models/email_account.dart';
import 'package:fasmail_app/config/api_config.dart';

class EmailAccountService {
  final ApiClient _client;

  EmailAccountService({required ApiClient client}) : _client = client;

  Future<List<EmailAccount>?> getAccounts() async {
    final response = await _client.get(ApiConfig.emailAccountsPath);
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body)['data'] as List;
      return data.map((a) => EmailAccount.fromJson(a)).toList();
    }
    return null;
  }

  Future<EmailAccount?> getAccount(String id) async {
    final response = await _client.get('${ApiConfig.emailAccountsPath}/$id');
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body)['data'];
      return EmailAccount.fromJson(data);
    }
    return null;
  }

  Future<EmailAccount?> createAccount(EmailAccount account) async {
    final response =
        await _client.post(ApiConfig.emailAccountsPath, account.toJson());
    if (response.statusCode == 201) {
      final data = jsonDecode(response.body)['data'];
      return EmailAccount.fromJson(data);
    }
    return null;
  }

  Future<bool> updateAccount(String id, EmailAccount account) async {
    final response = await _client.put(
        '${ApiConfig.emailAccountsPath}/$id', account.toJson());
    return response.statusCode == 200;
  }

  Future<bool> deleteAccount(String id) async {
    final response =
        await _client.delete('${ApiConfig.emailAccountsPath}/$id');
    return response.statusCode == 200;
  }
}
