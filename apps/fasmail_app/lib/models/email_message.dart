class EmailMessage {
  final int sequenceId;
  final String from;
  final String to;
  final String subject;
  final DateTime date;
  final String body;
  final bool isRead;
  final bool hasAttachments;

  EmailMessage({
    required this.sequenceId,
    required this.from,
    required this.to,
    required this.subject,
    required this.date,
    this.body = '',
    this.isRead = false,
    this.hasAttachments = false,
  });

  String get fromDisplay {
    final match = RegExp(r'"?([^"<]+)"?\s*<?').firstMatch(from);
    return match?.group(1)?.trim() ?? from;
  }

  String get preview {
    if (body.length <= 100) return body;
    return '${body.substring(0, 100)}...';
  }
}
