import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:fasmail_app/providers/auth_provider.dart';
import 'package:fasmail_app/screens/login_screen.dart';

class SetupScreen extends StatefulWidget {
  final String token;
  final String server;

  const SetupScreen({super.key, required this.token, required this.server});

  @override
  State<SetupScreen> createState() => _SetupScreenState();
}

class _SetupScreenState extends State<SetupScreen> {
  bool _isValidating = true;
  Map<String, dynamic>? _setupData;
  String? _error;

  @override
  void initState() {
    super.initState();
    _validateToken();
  }

  Future<void> _validateToken() async {
    final authProvider = context.read<AuthProvider>();
    final data =
        await authProvider.validateSetupToken(widget.token, widget.server);

    if (!mounted) return;

    setState(() {
      _isValidating = false;
      _setupData = data;
      _error = data == null ? 'Token inválido o expirado' : null;
    });
  }

  void _goToLogin() {
    Navigator.pushReplacement(
      context,
      MaterialPageRoute(builder: (_) => const LoginScreen()),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: _isValidating
              ? Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const CircularProgressIndicator(),
                    const SizedBox(height: 24),
                    Text(
                      'Verificando token de configuración...',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ],
                )
              : _error != null
                  ? Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.error_outline,
                            size: 64,
                            color: Theme.of(context).colorScheme.error),
                        const SizedBox(height: 16),
                        Text(_error!,
                            style: Theme.of(context)
                                .textTheme
                                .titleMedium
                                ?.copyWith(
                                    color:
                                        Theme.of(context).colorScheme.error)),
                        const SizedBox(height: 24),
                        FilledButton(
                          onPressed: _goToLogin,
                          child: const Text('Iniciar sesión manualmente'),
                        ),
                      ],
                    )
                  : Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.check_circle,
                            size: 64,
                            color: Theme.of(context).colorScheme.primary),
                        const SizedBox(height: 16),
                        Text(
                          'Cuenta configurada',
                          style: Theme.of(context)
                              .textTheme
                              .headlineSmall
                              ?.copyWith(fontWeight: FontWeight.bold),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Servidor: ${widget.server}',
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                        if (_setupData?['email'] != null)
                          Padding(
                            padding: const EdgeInsets.only(top: 4),
                            child: Text(
                              'Cuenta: ${_setupData!['email']}',
                              style: Theme.of(context).textTheme.bodyMedium,
                            ),
                          ),
                        if (_setupData?['email_accounts'] != null)
                          Padding(
                            padding: const EdgeInsets.only(top: 4),
                            child: Text(
                              '${(_setupData!['email_accounts'] as List).length} cuenta(s) de correo sincronizadas',
                              style: Theme.of(context).textTheme.bodySmall,
                            ),
                          ),
                        const SizedBox(height: 32),
                        FilledButton(
                          onPressed: _goToLogin,
                          child: const Text(
                              'Iniciar sesión para completar configuración'),
                        ),
                      ],
                    ),
        ),
      ),
    );
  }
}
