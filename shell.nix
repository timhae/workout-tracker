{
  pkgs ? import <nixpkgs> { },
}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    inotify-tools
    nodejs_24
    postgresql_17
  ];
}
