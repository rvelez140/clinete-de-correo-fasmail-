import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:fasmail_app/providers/admin_provider.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  @override
  void initState() {
    super.initState();
    context.read<AdminProvider>().loadSettings();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AdminProvider>(
      builder: (context, admin, _) {
        if (admin.isLoading && admin.settings == null) {
          return const Center(child: CircularProgressIndicator());
        }

        final settings = admin.settings ?? {};

        return RefreshIndicator(
          onRefresh: () => admin.loadSettings(),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text('Configuración del Sistema',
                  style: Theme.of(context).textTheme.headlineSmall),
              const SizedBox(height: 16),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(8),
                  child: settings.isEmpty
                      ? const Padding(
                          padding: EdgeInsets.all(16),
                          child: Text('No hay configuraciones'),
                        )
                      : Column(
                          children: settings.entries.map((entry) {
                            return ListTile(
                              title: Text(entry.key,
                                  style: const TextStyle(
                                      fontFamily: 'monospace',
                                      fontWeight: FontWeight.bold)),
                              subtitle: Text(entry.value),
                              dense: true,
                            );
                          }).toList(),
                        ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
