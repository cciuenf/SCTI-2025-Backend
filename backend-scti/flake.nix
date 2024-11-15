{
  description = "Rails API project with specified gem versions";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
  };

  outputs = { self, nixpkgs }: {
    devShells.x86_64-linux.default = let
      pkgs = import nixpkgs {
        system = "x86_64-linux";
      };
    in
      pkgs.mkShell {
        buildInputs = [
          pkgs.ruby_3_4           # Use a stable Ruby version
          pkgs.postgresql         # PostgreSQL client libraries
          pkgs.postgresql.dev     # Development headers for `pg` gem
          pkgs.libyaml            # Ensure libyaml is installed for psych gem
        ];

        shellHook = ''
          # Ensure Bundler is installed
          gem install bundler
          export GEM_HOME=$PWD/.gem
          export PATH=$GEM_HOME/bin:$PATH

          # Install the required gems with Bundler
          bundle install
        '';
      };
  };
}

