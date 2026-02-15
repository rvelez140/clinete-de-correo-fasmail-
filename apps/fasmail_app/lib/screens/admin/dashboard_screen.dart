import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:fasmail_app/providers/admin_provider.dart';

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  @override
  void initState() {
    super.initState();
    context.read<AdminProvider>().loadDashboard();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AdminProvider>(
      builder: (context, admin, _) {
        if (admin.isLoading && admin.stats == null) {
          return const Center(child: CircularProgressIndicator());
        }

        final stats = admin.stats;
        if (stats == null) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, size: 48),
                const SizedBox(height: 16),
                Text(admin.error ?? 'Error al cargar datos'),
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: () => admin.loadDashboard(),
                  child: const Text('Reintentar'),
                ),
              ],
            ),
          );
        }

        return RefreshIndicator(
          onRefresh: () => admin.loadDashboard(),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text('Dashboard',
                  style: Theme.of(context).textTheme.headlineSmall),
              const SizedBox(height: 16),
              _buildStatsGrid(context, stats),
              const SizedBox(height: 24),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Estado del Sistema',
                          style: Theme.of(context).textTheme.titleMedium),
                      const Divider(),
                      _buildStatusRow(
                          'PostgreSQL', stats.postgresOK, context),
                      _buildStatusRow('Redis', stats.redisOK, context),
                      if (stats.systemVersion.isNotEmpty)
                        ListTile(
                          leading: const Icon(Icons.info_outline),
                          title: const Text('Versión'),
                          trailing: Text(stats.systemVersion),
                        ),
                      if (stats.installedAt.isNotEmpty)
                        ListTile(
                          leading: const Icon(Icons.calendar_today),
                          title: const Text('Instalado'),
                          trailing: Text(stats.installedAt),
                        ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildStatsGrid(BuildContext context, stats) {
    return Wrap(
      spacing: 12,
      runSpacing: 12,
      children: [
        _buildStatCard(
            context, 'Usuarios', stats.totalUsers.toString(), Icons.people),
        _buildStatCard(context, 'Empresas', stats.totalCompanies.toString(),
            Icons.business),
      ],
    );
  }

  Widget _buildStatCard(
      BuildContext context, String label, String value, IconData icon) {
    return SizedBox(
      width: (MediaQuery.of(context).size.width - 44) / 2,
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: [
              Icon(icon, size: 32, color: Theme.of(context).colorScheme.primary),
              const SizedBox(height: 8),
              Text(value,
                  style: Theme.of(context)
                      .textTheme
                      .headlineMedium
                      ?.copyWith(fontWeight: FontWeight.bold)),
              Text(label, style: Theme.of(context).textTheme.bodySmall),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildStatusRow(String name, bool isOk, BuildContext context) {
    return ListTile(
      leading: Icon(
        isOk ? Icons.check_circle : Icons.error,
        color: isOk
            ? Theme.of(context).colorScheme.primary
            : Theme.of(context).colorScheme.error,
      ),
      title: Text(name),
      trailing: Text(isOk ? 'OK' : 'Error',
          style: TextStyle(
            color: isOk
                ? Theme.of(context).colorScheme.primary
                : Theme.of(context).colorScheme.error,
          )),
    );
  }
}
