#define MyAppName "muKsMaTT"
#ifndef MyAppVersion
  #define MyAppVersion "0.0.0"
#endif
#define MyAppExeName "muksmatt.exe"

[Setup]
AppId={{9DDE20AE-1B37-4CB7-9B91-282F913A39A8}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
DefaultDirName={autopf}\muKsMaTT
DefaultGroupName=muKsMaTT
DisableProgramGroupPage=yes
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=admin
OutputDir=..\..\dist
#ifdef ThinSetup
OutputBaseFilename=muKsMaTT-{#MyAppVersion}-Thin-Setup
#else
OutputBaseFilename=muKsMaTT-{#MyAppVersion}-Setup
#endif
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
UninstallDisplayIcon={app}\{#MyAppExeName}
VersionInfoVersion={#MyAppVersion}.0
VersionInfoProductName=muKsMaTT
VersionInfoDescription=muKsMaTT DVD remuxer

[Tasks]
Name: "desktopicon"; Description: "Create a desktop shortcut"; GroupDescription: "Additional shortcuts:"; Flags: unchecked

[Files]
Source: "..\..\muKsMaTT.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\muksmatt-cli.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\src\muksmatt.exe.manifest"; DestDir: "{app}"; DestName: "muksmatt.exe.manifest"; Flags: ignoreversion
Source: "..\..\src\muksmatt.ico"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\THIRD_PARTY.md"; DestDir: "{app}"; Flags: ignoreversion
#ifndef ThinSetup
Source: "bundled-tools\ffmpeg-2026-09-08\*"; DestDir: "{localappdata}\muKsMaTT\tools\ffmpeg-2026-09-08"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "bundled-tools\mediainfo-26.05\MediaInfo.exe"; DestDir: "{localappdata}\muKsMaTT\tools\mediainfo-26.05"; Flags: ignoreversion
#endif

[Icons]
Name: "{autoprograms}\muKsMaTT"; Filename: "{app}\{#MyAppExeName}"; WorkingDir: "{app}"
Name: "{autodesktop}\muKsMaTT"; Filename: "{app}\{#MyAppExeName}"; WorkingDir: "{app}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Launch muKsMaTT"; Flags: nowait postinstall skipifsilent