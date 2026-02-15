import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:fasmail_app/providers/admin_provider.dart';

class UsersScreen extends StatefulWidget {
  const UsersScreen({super.key});

  @override
  State<UsersScreen> createState() => _UsersScreenState();
}

class _UsersScreenState extends State<UsersScreen> {
  @override
  void initState() {
    super.initState();
    context.read<AdminProvider>().loadUsers();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Usuarios')),
      body: Consumer<AdminProvider>(
        builder: (context, admin, _) {
          if (admin.isLoading && admin.users == null) {
            return const Center(child: CircularProgressIndicator());
          }

          final users = admin.users ?? [];

          return RefreshIndicator(
            onRefresh: () => admin.loadUsers(),
            child: users.isEmpty
                ? const Center(child: Text('No hay usuarios'))
                : ListView.builder(
                    itemCount: users.length,
                    itemBuilder: (context, index) {
                      final user = users[index];
                      return ListTile(
                        leading: CircleAvatar(
                          child: Text(user.displayName.isNotEmpty
                              ? user.displayName[0].toUpperCase()
                              : user.email[0].toUpperCase()),
                        ),
                        title: Text(user.displayName.isNotEmpty
                            ? user.displayName
                            : user.email),
                        subtitle: Text('${user.email} - ${user.role}'),
                        trailing: user.isActive
                            ? const Icon(Icons.check_circle,
                                color: Colors.green, size: 20)
                            : const Icon(Icons.cancel,
                                color: Colors.red, size: 20),
                      );
                    },
                  ),
          );
        },
      ),
    );
  }
}
