import 'package:flutter/material.dart';
import 'package:fasmail_app/screens/splash_screen.dart';
import 'package:fasmail_app/screens/login_screen.dart';
import 'package:fasmail_app/screens/setup_screen.dart';
import 'package:fasmail_app/screens/home_screen.dart';

class AppRoutes {
  static const String splash = '/';
  static const String login = '/login';
  static const String setup = '/setup';
  static const String home = '/home';

  static Map<String, WidgetBuilder> get routes => {
        splash: (context) => const SplashScreen(),
        login: (context) => const LoginScreen(),
        home: (context) => const HomeScreen(),
      };
}
