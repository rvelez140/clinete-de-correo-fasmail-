class Company {
  final String id;
  final String name;
  final String slug;
  final String logoPath;
  final String primaryColor;
  final String successColor;
  final String dangerColor;
  final String warningColor;
  final bool isActive;

  Company({
    required this.id,
    required this.name,
    required this.slug,
    this.logoPath = '',
    this.primaryColor = '#2563eb',
    this.successColor = '#16a34a',
    this.dangerColor = '#dc2626',
    this.warningColor = '#d97706',
    this.isActive = true,
  });

  factory Company.fromJson(Map<String, dynamic> json) {
    return Company(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      slug: json['slug'] ?? '',
      logoPath: json['logo_path'] ?? '',
      primaryColor: json['primary_color'] ?? '#2563eb',
      successColor: json['success_color'] ?? '#16a34a',
      dangerColor: json['danger_color'] ?? '#dc2626',
      warningColor: json['warning_color'] ?? '#d97706',
      isActive: json['is_active'] ?? true,
    );
  }

  Map<String, dynamic> toJson() => {
        'name': name,
        'slug': slug,
        'primary_color': primaryColor,
        'success_color': successColor,
        'danger_color': dangerColor,
        'warning_color': warningColor,
        'is_active': isActive,
      };
}
