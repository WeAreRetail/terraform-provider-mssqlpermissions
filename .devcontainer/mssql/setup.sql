USE master
GO
IF NOT EXISTS(SELECT * FROM sys.databases WHERE name = 'ApplicationDB')
  BEGIN
    CREATE DATABASE [ApplicationDB]
  END
GO
