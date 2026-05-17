{
  description = "a minimal CLI for managing AI skills and a global CLAUDE.md";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachSystem [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ] (system:
      let
        pkgs = import nixpkgs { inherit system; };

        pname = "skill-cli";
        version = "1.1.0";

        skill-cli = pkgs.buildGoModule {
          inherit pname version;
          src = ./.;
          vendorHash = null;
          subPackages = [ "." ];
          env.CGO_ENABLED = "0";
          ldflags = [ "-s" "-w" ];
          doCheck = false;
        };
      in
      {
        packages.default = skill-cli;
        packages.skill-cli = skill-cli;

        apps.default = {
          type = "app";
          program = "${skill-cli}/bin/skill-cli";
        };
      });
}
