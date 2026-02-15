import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import 'package:fasmail_app/providers/email_provider.dart';
import 'package:fasmail_app/screens/email/compose_screen.dart';
import 'package:fasmail_app/screens/email/email_detail_screen.dart';
import 'package:fasmail_app/screens/email/account_settings_screen.dart';

class InboxScreen extends StatefulWidget {
  const InboxScreen({super.key});

  @override
  State<InboxScreen> createState() => _InboxScreenState();
}

class _InboxScreenState extends State<InboxScreen> {
  @override
  void initState() {
    super.initState();
    final emailProvider = context.read<EmailProvider>();
    emailProvider.loadAccounts();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<EmailProvider>(
      builder: (context, emailProvider, _) {
        // No accounts configured
        if (emailProvider.accounts != null && emailProvider.accounts!.isEmpty) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.mail_outline, size: 64, color: Colors.grey),
                const SizedBox(height: 16),
                const Text('No hay cuentas de correo configuradas'),
                const SizedBox(height: 16),
                FilledButton.icon(
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                          builder: (_) => const AccountSettingsScreen()),
                    );
                  },
                  icon: const Icon(Icons.add),
                  label: const Text('Agregar cuenta'),
                ),
              ],
            ),
          );
        }

        // Account selector + messages
        return Column(
          children: [
            // Account selector
            if (emailProvider.accounts != null &&
                emailProvider.accounts!.isNotEmpty)
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                child: Row(
                  children: [
                    Expanded(
                      child: DropdownButton<String>(
                        isExpanded: true,
                        value: emailProvider.activeAccount?.id,
                        hint: const Text('Seleccionar cuenta'),
                        items: emailProvider.accounts!.map((account) {
                          return DropdownMenuItem(
                            value: account.id,
                            child: Text(account.emailAddress),
                          );
                        }).toList(),
                        onChanged: (id) {
                          final account = emailProvider.accounts!
                              .firstWhere((a) => a.id == id);
                          emailProvider.connectToAccount(account);
                        },
                      ),
                    ),
                    if (emailProvider.isConnecting)
                      const Padding(
                        padding: EdgeInsets.only(left: 8),
                        child: SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                      ),
                  ],
                ),
              ),

            // Error message
            if (emailProvider.error != null)
              Container(
                margin: const EdgeInsets.symmetric(horizontal: 16),
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.errorContainer,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    Icon(Icons.error_outline,
                        color: Theme.of(context).colorScheme.error),
                    const SizedBox(width: 8),
                    Expanded(
                        child: Text(emailProvider.error!,
                            style: TextStyle(
                                color:
                                    Theme.of(context).colorScheme.error))),
                  ],
                ),
              ),

            // Messages
            Expanded(
              child: emailProvider.isConnecting
                  ? const Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          CircularProgressIndicator(),
                          SizedBox(height: 16),
                          Text('Conectando al servidor de correo...'),
                        ],
                      ),
                    )
                  : emailProvider.messages == null
                      ? const Center(
                          child: Text(
                              'Seleccione una cuenta para ver sus correos'))
                      : emailProvider.messages!.isEmpty
                          ? const Center(
                              child: Text('No hay mensajes en la bandeja'))
                          : RefreshIndicator(
                              onRefresh: () => emailProvider.fetchMessages(),
                              child: ListView.separated(
                                itemCount: emailProvider.messages!.length,
                                separatorBuilder: (_, __) =>
                                    const Divider(height: 1),
                                itemBuilder: (context, index) {
                                  final msg = emailProvider.messages![index];
                                  return ListTile(
                                    leading: CircleAvatar(
                                      child: Text(
                                        msg.fromDisplay.isNotEmpty
                                            ? msg.fromDisplay[0].toUpperCase()
                                            : '?',
                                      ),
                                    ),
                                    title: Text(
                                      msg.subject,
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: TextStyle(
                                        fontWeight: msg.isRead
                                            ? FontWeight.normal
                                            : FontWeight.bold,
                                      ),
                                    ),
                                    subtitle: Text(
                                      '${msg.fromDisplay}\n${msg.preview}',
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                    trailing: Column(
                                      mainAxisAlignment:
                                          MainAxisAlignment.center,
                                      children: [
                                        Text(
                                          DateFormat('dd/MM').format(msg.date),
                                          style: Theme.of(context)
                                              .textTheme
                                              .bodySmall,
                                        ),
                                        if (msg.hasAttachments)
                                          const Icon(Icons.attach_file,
                                              size: 16),
                                      ],
                                    ),
                                    isThreeLine: true,
                                    onTap: () {
                                      Navigator.push(
                                        context,
                                        MaterialPageRoute(
                                          builder: (_) =>
                                              EmailDetailScreen(message: msg),
                                        ),
                                      );
                                    },
                                  );
                                },
                              ),
                            ),
            ),
          ],
        );
      },
    );
  }
}
