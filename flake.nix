{
  description = "Another Clash Kernel";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/master";

  inputs.utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ self.overlay ];
          };
        in
        rec {
          packages.default = pkgs.clash-rev;
        }
      ) //
    (
      let version = nixpkgs.lib.substring 0 8 self.lastModifiedDate or self.lastModified or "19700101"; in
      {
        overlay = final: prev: {

          clash-rev = final.buildGo119Module {
            pname = "clash-rev";
            inherit version;
            src = ./.;

            vendorSha256 = "sha256-W5oiPtTRin0731QQWr98xZ2Vpk97HYcBtKoi1OKZz+w=";

            # Do not build testing suit
            excludedPackages = [ "./test" ];

            CGO_ENABLED = 0;

            ldflags = [
              "-s"
              "-w"
              "-X github.com/MerlinKodo/clash-rev/constant.Version=dev-${version}"
              "-X github.com/MerlinKodo/clash-rev/constant.BuildTime=${version}"
            ];
            
            tags = [
              "with_gvisor"
            ];

            # Network required 
            doCheck = false;

            postInstall = ''
              mv $out/bin/clash $out/bin/clash-rev
            '';

          };
        };
      }
    );
}

