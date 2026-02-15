import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:fasmail_app/providers/email_provider.dart';
import 'package:fasmail_app/models/email_account.dart';

class AccountSettingsScreen extends StatefulWidget {
  const AccountSettingsScreen({super.key});

  @override
  State<AccountSettingsScreen> createState() => _AccountSettingsScreenState();
}

class _AccountSettingsScreenState extends State<AccountSettingsScreen> {
  @override
  void initState() {
    super.initState();
    context.read<EmailProvider>().loadAccounts();
  }

  void _showAddAccountDialog() {
    final emailController = TextEditingController();
    final displayNameController = TextEditingController();
    final imapHostController = TextEditingController();
    final imapPortController = TextEditingController(text: '993');
    final smtpHostController = TextEditingController();
    final smtpPortController = TextEditingController(text: '587');
    final usernameController = TextEditingController();
    final passwordController = TextEditingController();

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Agregar cuenta de correo'),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: emailController,
                decoration: const InputDecoration(labelText: 'Email *'),
                keyboardType: TextInputType.emailAddress,
              ),
              TextField(
                controller: displayNameController,
                decoration: const InputDecoration(labelText: 'Nombre para mostrar'),
              ),
              const SizedBox(height: 8),
              const Text('Servidor IMAP', style: TextStyle(fontWeight: FontWeight.bold)),
              TextField(
                controller: imapHostController,
                decoration: const InputDecoration(labelText: 'Host IMAP *'),
              ),
              TextField(
                controller: imapPortController,
                decoration: const InputDecoration(labelText: 'Puerto IMAP'),
                keyboardType: TextInputType.number,
              ),
              const SizedBox(height: 8),
              const Text('Servidor SMTP', style: TextStyle(fontWeight: FontWeight.bold)),
              TextField(
                controller: smtpHostController,
                decoration: const InputDecoration(labelText: 'Host SMTP *'),
              ),
              TextField(
                controller: smtpPortController,
                decoration: const InputDecoration(labelText: 'Puerto SMTP'),
                keyboardType: TextInputType.number,
              ),
              const SizedBox(height: 8),
              const Text('Credenciales', style: TextStyle(fontWeight: FontWeight.bold)),
              TextField(
                controller: usernameController,
                decoration: const InputDecoration(labelText: 'Usuario *'),
              ),
              TextField(
                controller: passwordController,
                decoration: const InputDecoration(labelText: 'Contraseña *'),
                obscureText: true,
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancelar'),
          ),
          FilledButton(
            onPressed: () async {
              if (emailController.text.isEmpty ||
                  imapHostController.text.isEmpty ||
                  smtpHostController.text.isEmpty ||
                  usernameController.text.isEmpty ||
                  passwordController.text.isEmpty) {
                return;
              }

              final account = EmailAccount(
                emailAddress: emailController.text.trim(),
                displayName: displayNameController.text.trim(),
                imapHost: imapHostController.text.trim(),
                imapPort: int.tryParse(imapPortController.text) ?? 993,
                smtpHost: smtpHostController.text.trim(),
                smtpPort: int.tryParse(smtpPortController.text) ?? 587,
                username: usernameController.text.trim(),
                password: passwordController.text,
              );

              final provider = context.read<EmailProvider>();
              final success = await provider.addAccount(account);

              if (success && mounted) {
                Navigator.pop(context);
              }
            },
            child: const Text('Agregar'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Cuentas de correo')),
      body: Consumer<EmailProvider>(
        builder: (context, provider, _) {
          if (provider.isLoading && provider.accounts == null) {
            return const Center(child: CircularProgressIndicator());
          }

          final accounts = provider.accounts ?? [];

          return accounts.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.email_outlined, size: 48, color: Colors.grey),
                      const SizedBox(height: 16),
                      const Text('No hay cuentas configuradas'),
                      const SizedBox(height: 16),
                      FilledButton.icon(
                        onPressed: _showAddAccountDialog,
                        icon: const Icon(Icons.add),
                        label: const Text('Agregar cuenta'),
                      ),
                    ],
                  ),
                )
              : ListView.builder(
                  itemCount: accounts.length,
                  itemBuilder: (context, index) {
                    final account = accounts[index];
                    return ListTile(
                      leading: CircleAvatar(
                        child: Text(account.emailAddress[0].toUpperCase()),
                      ),
                      title: Text(account.emailAddress),
                      subtitle: Text(
                        '${account.imapHost}:${account.imapPort}',
                      ),
                      trailing: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          if (account.isDefault)
                            const Chip(label: Text('Principal')),
                          IconButton(
                            icon: const Icon(Icons.delete_outline),
                            onPressed: () async {
                              final confirm = await showDialog<bool>(
                                context: context,
                                builder: (ctx) => AlertDialog(
                                  title: const Text('Eliminar cuenta'),
                                  content: Text(
                                      'Eliminar ${account.emailAddress}?'),
                                  actions: [
                                    TextButton(
                                      onPressed: () =>
                                          Navigator.pop(ctx, false),
                                      child: const Text('Cancelar'),
                                    ),
                                    FilledButton(
                                      onPressed: () =>
                                          Navigator.pop(ctx, true),
                                      child: const Text('Eliminar'),
                                    ),
                                  ],
                                ),
                              );
                              if (confirm == true) {
                                provider.deleteAccount(account.id);
                              }
                            },
                          ),
                        ],
                      ),
                    );
                  },
                );
        },
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _showAddAccountDialog,
        child: const Icon(Icons.add),
      ),
    );
  }
}
