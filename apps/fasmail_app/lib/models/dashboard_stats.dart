class DashboardStats {
  final int totalUsers;
  final int activeUsers;
  final int totalCompanies;
  final String systemVersion;
  final String installedAt;
  final bool postgresOK;
  final bool redisOK;

  DashboardStats({
    this.totalUsers = 0,
    this.activeUsers = 0,
    this.totalCompanies = 0,
    this.systemVersion = '',
    this.installedAt = '',
    this.postgresOK = false,
    this.redisOK = false,
  });

  factory DashboardStats.fromJson(Map<String, dynamic> json) {
    return DashboardStats(
      totalUsers: json['TotalUsers'] ?? json['total_users'] ?? 0,
      activeUsers: json['ActiveUsers'] ?? json['active_users'] ?? 0,
      totalCompanies: json['TotalCompanies'] ?? json['total_companies'] ?? 0,
      systemVersion: json['SystemVersion'] ?? json['system_version'] ?? '',
      installedAt: json['InstalledAt'] ?? json['installed_at'] ?? '',
      postgresOK: json['PostgresOK'] ?? json['postgres_ok'] ?? false,
      redisOK: json['RedisOK'] ?? json['redis_ok'] ?? false,
    );
  }
}
