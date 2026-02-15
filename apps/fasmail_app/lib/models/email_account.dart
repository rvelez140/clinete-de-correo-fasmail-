class EmailAccount {
  final String id;
  final String emailAddress;
  final String displayName;
  final String imapHost;
  final int imapPort;
  final bool imapUseTls;
  final String smtpHost;
  final int smtpPort;
  final bool smtpUseTls;
  final String username;
  final String? password;
  final bool isDefault;

  EmailAccount({
    this.id = '',
    required this.emailAddress,
    this.displayName = '',
    required this.imapHost,
    this.imapPort = 993,
    this.imapUseTls = true,
    required this.smtpHost,
    this.smtpPort = 587,
    this.smtpUseTls = true,
    required this.username,
    this.password,
    this.isDefault = false,
  });

  factory EmailAccount.fromJson(Map<String, dynamic> json) {
    return EmailAccount(
      id: json['id'] ?? '',
      emailAddress: json['email_address'] ?? '',
      displayName: json['display_name'] ?? '',
      imapHost: json['imap_host'] ?? '',
      imapPort: json['imap_port'] ?? 993,
      imapUseTls: json['imap_use_tls'] ?? true,
      smtpHost: json['smtp_host'] ?? '',
      smtpPort: json['smtp_port'] ?? 587,
      smtpUseTls: json['smtp_use_tls'] ?? true,
      username: json['username'] ?? '',
      password: json['password'],
      isDefault: json['is_default'] ?? false,
    );
  }

  Map<String, dynamic> toJson() => {
        'email_address': emailAddress,
        'display_name': displayName,
        'imap_host': imapHost,
        'imap_port': imapPort,
        'imap_use_tls': imapUseTls,
        'smtp_host': smtpHost,
        'smtp_port': smtpPort,
        'smtp_use_tls': smtpUseTls,
        'username': username,
        if (password != null) 'password': password,
        'is_default': isDefault,
      };
}
