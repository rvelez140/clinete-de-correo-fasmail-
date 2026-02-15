import 'package:enough_mail/enough_mail.dart';
import 'package:fasmail_app/models/email_account.dart';
import 'package:fasmail_app/models/email_message.dart';

class ImapService {
  MailClient? _mailClient;
  bool _isConnected = false;

  bool get isConnected => _isConnected;

  Future<void> connect(EmailAccount account) async {
    if (account.password == null || account.password!.isEmpty) {
      throw Exception('Contraseña de correo no disponible');
    }

    final mailAccount = MailAccount.fromManualSettingsWithAuth(
      name: account.displayName.isNotEmpty
          ? account.displayName
          : account.emailAddress,
      email: account.emailAddress,
      userName: account.username,
      password: account.password!,
      incomingHost: account.imapHost,
      incomingPort: account.imapPort,
      incomingTypeName: account.imapUseTls ? 'imaps' : 'imap',
      outgoingHost: account.smtpHost,
      outgoingPort: account.smtpPort,
      outgoingTypeName: account.smtpUseTls ? 'smtps' : 'smtp',
    );

    _mailClient = MailClient(mailAccount, isLogEnabled: false);
    await _mailClient!.connect();
    _isConnected = true;
  }

  Future<void> disconnect() async {
    if (_mailClient != null) {
      await _mailClient!.disconnect();
      _isConnected = false;
      _mailClient = null;
    }
  }

  Future<List<EmailMessage>> fetchInbox({int count = 30}) async {
    if (_mailClient == null || !_isConnected) {
      throw Exception('No conectado al servidor IMAP');
    }

    await _mailClient!.selectInbox();
    final messages = await _mailClient!.fetchMessages(count: count);

    return messages.map((msg) {
      return EmailMessage(
        sequenceId: msg.sequenceId ?? 0,
        from: msg.from?.first.toString() ?? '',
        to: msg.to?.first.toString() ?? '',
        subject: msg.decodeSubject() ?? '(Sin asunto)',
        date: msg.decodeDate() ?? DateTime.now(),
        body: msg.decodeTextPlainPart() ?? msg.decodeTextHtmlPart() ?? '',
        isRead: msg.isSeen,
        hasAttachments: msg.hasAttachments(),
      );
    }).toList();
  }

  Future<void> sendMessage({
    required EmailAccount account,
    required String to,
    required String subject,
    required String body,
  }) async {
    if (account.password == null || account.password!.isEmpty) {
      throw Exception('Contraseña de correo no disponible');
    }

    final smtpClient = SmtpClient(
      account.smtpHost,
      port: account.smtpPort,
      isLogEnabled: false,
    );

    if (account.smtpUseTls) {
      await smtpClient.connectToServer(
        account.smtpHost,
        account.smtpPort,
        isSecure: true,
      );
    } else {
      await smtpClient.connectToServer(
        account.smtpHost,
        account.smtpPort,
        isSecure: false,
      );
    }

    await smtpClient.ehlo();

    if (!account.smtpUseTls) {
      await smtpClient.startTls();
    }

    await smtpClient.authenticate(account.username, account.password!);

    final message = MessageBuilder.prepareMultipartAlternativeMessage(
      plainText: body,
      htmlText: '<p>${body.replaceAll('\n', '<br>')}</p>',
    );
    message.from = [MailAddress(account.displayName, account.emailAddress)];
    message.to = [MailAddress(null, to)];
    message.subject = subject;

    await smtpClient.sendMessage(message.buildMimeMessage());
    await smtpClient.closeConnection();
  }
}
