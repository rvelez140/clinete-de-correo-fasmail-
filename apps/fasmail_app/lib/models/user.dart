class User {
  final String id;
  final String email;
  final String displayName;
  final String role;
  final bool mustChangePassword;
  final bool isActive;
  final String? companyId;
  final DateTime? lastLoginAt;
  final DateTime? createdAt;

  User({
    required this.id,
    required this.email,
    required this.displayName,
    required this.role,
    this.mustChangePassword = false,
    this.isActive = true,
    this.companyId,
    this.lastLoginAt,
    this.createdAt,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] ?? '',
      email: json['email'] ?? '',
      displayName: json['display_name'] ?? '',
      role: json['role'] ?? 'user',
      mustChangePassword: json['must_change_password'] ?? false,
      isActive: json['is_active'] ?? true,
      companyId: json['company_id'],
      lastLoginAt: json['last_login_at'] != null
          ? DateTime.tryParse(json['last_login_at'])
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'])
          : null,
    );
  }

  bool get isSuperAdmin => role == 'super_admin';
  bool get isAdmin => role == 'admin' || role == 'super_admin';
}
