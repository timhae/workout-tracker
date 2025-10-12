{
  pkgs ? import <nixpkgs> { },
}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    fd
    go
    gopls
    nodejs_24
    postgresql_17
    wgo
  ];
}
