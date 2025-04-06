{
  description = "SCTI 2025";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        config.allowUnfree = true;
      };
    in {
      devShell = pkgs.mkShell {
        buildInputs = with pkgs; [
          postman
          go
          docker
          docker-compose
        ];
        shellHook = ''
          echo "Welcome to the development environment!"
          export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath [ pkgs.libglvnd ]}:$LD_LIBRARY_PATH
          
          # Verificar se o Docker está rodando
          if ! docker info > /dev/null 2>&1; then
            echo "⚠️  Atenção: Docker não está rodando. Por favor, inicie o serviço Docker."
          else
            echo "✅ Docker está rodando!"
          fi
          
          USER_SHELL=$(getent passwd $USER | cut -d: -f7)
          exec $USER_SHELL
        '';
      };
    }
  );
}
