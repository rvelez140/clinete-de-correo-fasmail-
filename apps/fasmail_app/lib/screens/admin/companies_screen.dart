import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:fasmail_app/providers/admin_provider.dart';
import 'package:fasmail_app/screens/admin/company_form_screen.dart';

class CompaniesScreen extends StatefulWidget {
  const CompaniesScreen({super.key});

  @override
  State<CompaniesScreen> createState() => _CompaniesScreenState();
}

class _CompaniesScreenState extends State<CompaniesScreen> {
  @override
  void initState() {
    super.initState();
    context.read<AdminProvider>().loadCompanies();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<AdminProvider>(
      builder: (context, admin, _) {
        if (admin.isLoading && admin.companies == null) {
          return const Center(child: CircularProgressIndicator());
        }

        final companies = admin.companies ?? [];

        return Scaffold(
          body: RefreshIndicator(
            onRefresh: () => admin.loadCompanies(),
            child: companies.isEmpty
                ? const Center(child: Text('No hay empresas'))
                : ListView.builder(
                    itemCount: companies.length,
                    itemBuilder: (context, index) {
                      final company = companies[index];
                      return ListTile(
                        leading: CircleAvatar(
                          backgroundColor: _parseColor(company.primaryColor),
                          child: Text(
                            company.name.isNotEmpty
                                ? company.name[0].toUpperCase()
                                : '?',
                            style: const TextStyle(color: Colors.white),
                          ),
                        ),
                        title: Text(company.name),
                        subtitle: Text(company.slug),
                        trailing: company.isActive
                            ? const Icon(Icons.check_circle,
                                color: Colors.green, size: 20)
                            : const Icon(Icons.cancel,
                                color: Colors.red, size: 20),
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (_) =>
                                  CompanyFormScreen(company: company),
                            ),
                          );
                        },
                      );
                    },
                  ),
          ),
          floatingActionButton: FloatingActionButton(
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => const CompanyFormScreen(),
                ),
              );
            },
            child: const Icon(Icons.add),
          ),
        );
      },
    );
  }

  Color _parseColor(String hex) {
    try {
      return Color(int.parse(hex.replaceFirst('#', 'FF'), radix: 16));
    } catch (_) {
      return Colors.blue;
    }
  }
}
