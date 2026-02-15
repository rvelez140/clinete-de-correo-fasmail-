import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:fasmail_app/models/email_message.dart';
import 'package:fasmail_app/screens/email/compose_screen.dart';

class EmailDetailScreen extends StatelessWidget {
  final EmailMessage message;

  const EmailDetailScreen({super.key, required this.message});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Mensaje'),
        actions: [
          IconButton(
            icon: const Icon(Icons.reply),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => ComposeScreen(
                    replyTo: message.from,
                    subject: 'Re: ${message.subject}',
                  ),
                ),
              );
            },
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              message.subject,
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                CircleAvatar(
                  child: Text(message.fromDisplay.isNotEmpty
                      ? message.fromDisplay[0].toUpperCase()
                      : '?'),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(message.fromDisplay,
                          style: const TextStyle(fontWeight: FontWeight.bold)),
                      Text(message.from,
                          style: Theme.of(context).textTheme.bodySmall),
                    ],
                  ),
                ),
                Text(
                  DateFormat('dd/MM/yyyy HH:mm').format(message.date),
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ),
            if (message.to.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text('Para: ${message.to}',
                  style: Theme.of(context).textTheme.bodySmall),
            ],
            const Divider(height: 32),
            SelectableText(
              message.body,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
          ],
        ),
      ),
    );
  }
}
