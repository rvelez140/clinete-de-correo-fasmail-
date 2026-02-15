import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:fasmail_app/providers/auth_provider.dart';
import 'package:fasmail_app/providers/admin_provider.dart';
import 'package:fasmail_app/providers/email_provider.dart';
import 'package:fasmail_app/screens/email/inbox_screen.dart';
import 'package:fasmail_app/screens/admin/dashboard_screen.dart';
import 'package:fasmail_app/screens/admin/companies_screen.dart';
import 'package:fasmail_app/screens/admin/settings_screen.dart';
import 'package:fasmail_app/screens/admin/users_screen.dart';
import 'package:fasmail_app/screens/email/account_settings_screen.dart';
import 'package:fasmail_app/screens/login_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _selectedIndex = 0;

  @override
  void initState() {
    super.initState();
    final authProvider = context.read<AuthProvider>();
    context.read<AdminProvider>().init(authProvider.authService);
    context.read<EmailProvider>().init(authProvider.authService);
  }

  @override
  Widget build(BuildContext context) {
    final authProvider = context.watch<AuthProvider>();
    final user = authProvider.currentUser;

    final screens = <Widget>[
      const InboxScreen(),
      const DashboardScreen(),
      if (user?.isSuperAdmin == true) const CompaniesScreen(),
      const SettingsScreen(),
    ];

    final destinations = <NavigationDestination>[
      const NavigationDestination(
        icon: Icon(Icons.inbox_outlined),
        selectedIcon: Icon(Icons.inbox),
        label: 'Correo',
      ),
      const NavigationDestination(
        icon: Icon(Icons.dashboard_outlined),
        selectedIcon: Icon(Icons.dashboard),
        label: 'Dashboard',
      ),
      if (user?.isSuperAdmin == true)
        const NavigationDestination(
          icon: Icon(Icons.business_outlined),
          selectedIcon: Icon(Icons.business),
          label: 'Empresas',
        ),
      const NavigationDestination(
        icon: Icon(Icons.settings_outlined),
        selectedIcon: Icon(Icons.settings),
        label: 'Ajustes',
      ),
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('FasMail'),
        actions: [
          if (user != null)
            PopupMenuButton<String>(
              icon: const Icon(Icons.account_circle_outlined),
              onSelected: (value) async {
                if (value == 'logout') {
                  await authProvider.logout();
                  if (mounted) {
                    Navigator.pushReplacement(
                      context,
                      MaterialPageRoute(builder: (_) => const LoginScreen()),
                    );
                  }
                } else if (value == 'accounts') {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                        builder: (_) => const AccountSettingsScreen()),
                  );
                } else if (value == 'users') {
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (_) => const UsersScreen()),
                  );
                }
              },
              itemBuilder: (context) => [
                PopupMenuItem(
                  enabled: false,
                  child: Text(user.email,
                      style: const TextStyle(fontWeight: FontWeight.bold)),
                ),
                const PopupMenuDivider(),
                const PopupMenuItem(
                  value: 'accounts',
                  child: ListTile(
                    leading: Icon(Icons.email_outlined),
                    title: Text('Cuentas de correo'),
                    contentPadding: EdgeInsets.zero,
                  ),
                ),
                if (user.isAdmin)
                  const PopupMenuItem(
                    value: 'users',
                    child: ListTile(
                      leading: Icon(Icons.people_outlined),
                      title: Text('Usuarios'),
                      contentPadding: EdgeInsets.zero,
                    ),
                  ),
                const PopupMenuDivider(),
                const PopupMenuItem(
                  value: 'logout',
                  child: ListTile(
                    leading: Icon(Icons.logout),
                    title: Text('Cerrar sesión'),
                    contentPadding: EdgeInsets.zero,
                  ),
                ),
              ],
            ),
        ],
      ),
      body: IndexedStack(
        index: _selectedIndex,
        children: screens,
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _selectedIndex,
        onDestinationSelected: (index) {
          setState(() => _selectedIndex = index);
        },
        destinations: destinations,
      ),
    );
  }
}
