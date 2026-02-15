import 'package:flutter/material.dart';
import 'package:fasmail_app/services/api_client.dart';
import 'package:fasmail_app/services/email_service.dart';
import 'package:fasmail_app/services/imap_service.dart';
import 'package:fasmail_app/services/auth_service.dart';
import 'package:fasmail_app/models/email_account.dart';
import 'package:fasmail_app/models/email_message.dart';

class EmailProvider extends ChangeNotifier {
  EmailAccountService? _accountService;
  final ImapService _imapService = ImapService();

  List<EmailAccount>? _accounts;
  EmailAccount? _activeAccount;
  List<EmailMessage>? _messages;
  bool _isLoading = false;
  bool _isConnecting = false;
  String? _error;

  List<EmailAccount>? get accounts => _accounts;
  EmailAccount? get activeAccount => _activeAccount;
  List<EmailMessage>? get messages => _messages;
  bool get isLoading => _isLoading;
  bool get isConnecting => _isConnecting;
  String? get error => _error;
  bool get isConnected => _imapService.isConnected;

  void init(AuthService authService) {
    final client = ApiClient(authService: authService);
    _accountService = EmailAccountService(client: client);
  }

  Future<void> loadAccounts() async {
    if (_accountService == null) return;
    _isLoading = true;
    _error = null;
    notifyListeners();

    _accounts = await _accountService!.getAccounts();

    _isLoading = false;
    notifyListeners();
  }

  Future<void> connectToAccount(EmailAccount account) async {
    _isConnecting = true;
    _error = null;
    _activeAccount = account;
    notifyListeners();

    try {
      // Get account with decrypted password
      final fullAccount = await _accountService!.getAccount(account.id);
      if (fullAccount == null) {
        _error = 'No se pudo obtener la configuración de la cuenta';
        _isConnecting = false;
        notifyListeners();
        return;
      }

      await _imapService.disconnect();
      await _imapService.connect(fullAccount);
      _activeAccount = fullAccount;
      await fetchMessages();
    } catch (e) {
      _error = 'Error al conectar: $e';
    }

    _isConnecting = false;
    notifyListeners();
  }

  Future<void> fetchMessages() async {
    if (!_imapService.isConnected) return;
    _isLoading = true;
    notifyListeners();

    try {
      _messages = await _imapService.fetchInbox();
    } catch (e) {
      _error = 'Error al cargar mensajes: $e';
    }

    _isLoading = false;
    notifyListeners();
  }

  Future<bool> sendMessage({
    required String to,
    required String subject,
    required String body,
  }) async {
    if (_activeAccount == null) return false;

    try {
      await _imapService.sendMessage(
        account: _activeAccount!,
        to: to,
        subject: subject,
        body: body,
      );
      return true;
    } catch (e) {
      _error = 'Error al enviar: $e';
      notifyListeners();
      return false;
    }
  }

  Future<bool> addAccount(EmailAccount account) async {
    if (_accountService == null) return false;
    final result = await _accountService!.createAccount(account);
    if (result != null) {
      await loadAccounts();
      return true;
    }
    return false;
  }

  Future<bool> deleteAccount(String id) async {
    if (_accountService == null) return false;
    final success = await _accountService!.deleteAccount(id);
    if (success) {
      if (_activeAccount?.id == id) {
        await _imapService.disconnect();
        _activeAccount = null;
        _messages = null;
      }
      await loadAccounts();
    }
    return success;
  }

  @override
  void dispose() {
    _imapService.disconnect();
    super.dispose();
  }
}
