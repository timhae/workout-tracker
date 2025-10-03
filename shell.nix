{
  pkgs ? import <nixpkgs> { },
}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    inotify-tools
    nodejs_24
  ];
}
