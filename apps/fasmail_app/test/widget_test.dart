import 'package:flutter_test/flutter_test.dart';
import 'package:fasmail_app/models/user.dart';
import 'package:fasmail_app/models/company.dart';
import 'package:fasmail_app/models/email_account.dart';
import 'package:fasmail_app/models/dashboard_stats.dart';
import 'package:fasmail_app/models/email_message.dart';
import 'package:fasmail_app/config/api_config.dart';

void main() {
  group('User model', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 1,
        'email': 'admin@test.com',
        'name': 'Admin User',
        'role': 'super_admin',
        'company_id': 1,
        'must_change_password': false,
      };
      final user = User.fromJson(json);
      expect(user.id, 1);
      expect(user.email, 'admin@test.com');
      expect(user.name, 'Admin User');
      expect(user.isSuperAdmin, true);
      expect(user.isAdmin, false);
    });
  });

  group('Company model', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 1,
        'name': 'Test Co',
        'domain': 'test.com',
        'max_users': 10,
        'is_active': true,
      };
      final company = Company.fromJson(json);
      expect(company.id, 1);
      expect(company.name, 'Test Co');
      expect(company.domain, 'test.com');
    });

    test('toJson serializes correctly', () {
      final company = Company(
        id: 1,
        name: 'Test Co',
        domain: 'test.com',
        maxUsers: 10,
        isActive: true,
      );
      final json = company.toJson();
      expect(json['name'], 'Test Co');
      expect(json['domain'], 'test.com');
      expect(json['max_users'], 10);
    });
  });

  group('EmailAccount model', () {
    test('fromJson parses correctly', () {
      final json = {
        'id': 1,
        'user_id': 1,
        'email_address': 'user@mail.com',
        'display_name': 'User',
        'imap_host': 'imap.mail.com',
        'imap_port': 993,
        'imap_use_tls': true,
        'smtp_host': 'smtp.mail.com',
        'smtp_port': 465,
        'smtp_use_tls': true,
        'username': 'user@mail.com',
        'is_default': true,
      };
      final account = EmailAccount.fromJson(json);
      expect(account.emailAddress, 'user@mail.com');
      expect(account.imapHost, 'imap.mail.com');
      expect(account.isDefault, true);
    });
  });

  group('DashboardStats model', () {
    test('fromJson parses snake_case', () {
      final json = {
        'total_companies': 5,
        'total_users': 50,
        'active_sessions': 10,
      };
      final stats = DashboardStats.fromJson(json);
      expect(stats.totalCompanies, 5);
      expect(stats.totalUsers, 50);
      expect(stats.activeSessions, 10);
    });
  });

  group('EmailMessage model', () {
    test('fromDisplay returns sender name', () {
      final msg = EmailMessage(
        uid: 1,
        from: 'John Doe <john@test.com>',
        to: 'jane@test.com',
        subject: 'Hello',
        body: 'World content here',
        date: DateTime(2025, 1, 1),
      );
      expect(msg.fromDisplay, 'John Doe');
      expect(msg.preview, 'World content here');
    });

    test('fromDisplay returns email when no name', () {
      final msg = EmailMessage(
        uid: 2,
        from: 'test@test.com',
        to: 'jane@test.com',
        subject: 'Test',
        body: '',
        date: DateTime(2025, 1, 1),
      );
      expect(msg.fromDisplay, 'test@test.com');
    });
  });

  group('ApiConfig', () {
    test('paths are correct', () {
      expect(ApiConfig.login, '/api/v1/auth/login');
      expect(ApiConfig.dashboard, '/api/v1/admin/dashboard');
      expect(ApiConfig.companies, '/api/v1/admin/companies');
      expect(ApiConfig.emailAccounts, '/api/v1/email/accounts');
    });

    test('apiUrl constructs URL correctly', () {
      final url = apiUrl('https://fasmail.example.com', '/api/v1/auth/login');
      expect(url, 'https://fasmail.example.com/api/v1/auth/login');
    });
  });
}
