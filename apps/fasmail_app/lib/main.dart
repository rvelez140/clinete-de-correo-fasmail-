import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:fasmail_app/providers/auth_provider.dart';
import 'package:fasmail_app/providers/admin_provider.dart';
import 'package:fasmail_app/providers/email_provider.dart';
import 'package:fasmail_app/app.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthProvider()),
        ChangeNotifierProvider(create: (_) => AdminProvider()),
        ChangeNotifierProvider(create: (_) => EmailProvider()),
      ],
      child: const FasMailApp(),
    ),
  );
}
