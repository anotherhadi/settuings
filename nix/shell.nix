{
  pkgs,
  gitHooksLib,
  gomod2nixPkgs,
}: let
  hooks = gitHooksLib.run {
    src = ../.;
    hooks = {
      gofmt.enable = true;
      govet.enable = true;

      gomod2nix = {
        enable = true;
        name = "gomod2nix";
        entry = "gomod2nix --outdir ./nix";
        language = "system";
        files = "go\\.(mod|sum)$";
        pass_filenames = false;
      };

      inject-exec = {
        enable = true;
        name = "inject-exec";
        entry = "python3 .github/scripts/inject-exec.py README.md";
        language = "system";
        files = "(README\\.md|cmd/)";
        pass_filenames = false;
      };

      doctoc = {
        enable = true;
        name = "doctoc";
        entry = "doctoc --notitle README.md";
        language = "system";
        files = "(README\\.md|cmd/)";
        pass_filenames = false;
      };
    };
  };
in
  pkgs.mkShell {
    packages = with pkgs;
      [
        go
        gosec
        govulncheck
        python3
        doctoc
        trufflehog
        gomod2nixPkgs.gomod2nix
      ]
      ++ hooks.enabledPackages;

    shellHook = hooks.shellHook;
  }
