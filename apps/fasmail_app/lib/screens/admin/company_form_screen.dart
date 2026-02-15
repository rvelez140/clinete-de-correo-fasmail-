import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:fasmail_app/providers/admin_provider.dart';
import 'package:fasmail_app/models/company.dart';

class CompanyFormScreen extends StatefulWidget {
  final Company? company;

  const CompanyFormScreen({super.key, this.company});

  @override
  State<CompanyFormScreen> createState() => _CompanyFormScreenState();
}

class _CompanyFormScreenState extends State<CompanyFormScreen> {
  final _formKey = GlobalKey<FormState>();
  late TextEditingController _nameController;
  late TextEditingController _slugController;
  late TextEditingController _primaryColorController;
  late TextEditingController _successColorController;
  late TextEditingController _dangerColorController;
  late TextEditingController _warningColorController;
  bool _isLoading = false;

  bool get isEditing => widget.company != null;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.company?.name ?? '');
    _slugController = TextEditingController(text: widget.company?.slug ?? '');
    _primaryColorController =
        TextEditingController(text: widget.company?.primaryColor ?? '#2563eb');
    _successColorController =
        TextEditingController(text: widget.company?.successColor ?? '#16a34a');
    _dangerColorController =
        TextEditingController(text: widget.company?.dangerColor ?? '#dc2626');
    _warningColorController =
        TextEditingController(text: widget.company?.warningColor ?? '#d97706');
  }

  @override
  void dispose() {
    _nameController.dispose();
    _slugController.dispose();
    _primaryColorController.dispose();
    _successColorController.dispose();
    _dangerColorController.dispose();
    _warningColorController.dispose();
    super.dispose();
  }

  Future<void> _handleSubmit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    final company = Company(
      id: widget.company?.id ?? '',
      name: _nameController.text.trim(),
      slug: _slugController.text.trim(),
      primaryColor: _primaryColorController.text.trim(),
      successColor: _successColorController.text.trim(),
      dangerColor: _dangerColorController.text.trim(),
      warningColor: _warningColorController.text.trim(),
    );

    final admin = context.read<AdminProvider>();
    final success = await admin.createCompany(company);

    setState(() => _isLoading = false);

    if (success && mounted) {
      Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(isEditing ? 'Editar Empresa' : 'Nueva Empresa'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: Column(
            children: [
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(
                  labelText: 'Nombre',
                  border: OutlineInputBorder(),
                ),
                validator: (v) =>
                    v?.isEmpty == true ? 'Nombre requerido' : null,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _slugController,
                decoration: const InputDecoration(
                  labelText: 'Slug',
                  helperText: 'Solo letras minúsculas, números y guiones',
                  border: OutlineInputBorder(),
                ),
                validator: (v) =>
                    v?.isEmpty == true ? 'Slug requerido' : null,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _primaryColorController,
                decoration: const InputDecoration(
                  labelText: 'Color primario',
                  hintText: '#2563eb',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _successColorController,
                decoration: const InputDecoration(
                  labelText: 'Color éxito',
                  hintText: '#16a34a',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _dangerColorController,
                decoration: const InputDecoration(
                  labelText: 'Color peligro',
                  hintText: '#dc2626',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _warningColorController,
                decoration: const InputDecoration(
                  labelText: 'Color advertencia',
                  hintText: '#d97706',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                child: FilledButton(
                  onPressed: _isLoading ? null : _handleSubmit,
                  child: _isLoading
                      ? const CircularProgressIndicator(strokeWidth: 2)
                      : Text(isEditing ? 'Guardar' : 'Crear'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
