class CodacyCliV2 < Formula
  version "1.0.0"
  url "https://raw.githubusercontent.com/codacy/codacy-cli-v2/8ce8c5da6ee6cf3ff145b7b18f861a6d740e11c0/codacy-cli.sh"
  sha256 "fb616e2f5c639985566c81a6e6ce51db2e8de56bf217e837d13efe2f3ccc3042"

  def install
    bin.install "codacy-cli.sh" => "codacy-cli"
  end
end
