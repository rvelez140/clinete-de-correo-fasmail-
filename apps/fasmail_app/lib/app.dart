import 'package:flutter/material.dart';
import 'package:fasmail_app/config/routes.dart';
import 'package:fasmail_app/screens/splash_screen.dart';

class FasMailApp extends StatelessWidget {
  const FasMailApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'FasMail',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorSchemeSeed: const Color(0xFF2563EB),
        useMaterial3: true,
        brightness: Brightness.light,
      ),
      darkTheme: ThemeData(
        colorSchemeSeed: const Color(0xFF2563EB),
        useMaterial3: true,
        brightness: Brightness.dark,
      ),
      themeMode: ThemeMode.system,
      routes: AppRoutes.routes,
      home: const SplashScreen(),
    );
  }
}
