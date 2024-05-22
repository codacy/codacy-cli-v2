class CodacyCliV2 < Formula
  version "0.1.0-main.35.70f0f48"
  url "https://raw.githubusercontent.com/codacy/codacy-cli-v2/0.1.0-main.35.70f0f48/codacy-cli.sh"
  sha256 "fb616e2f5c639985566c81a6e6ce51db2e8de56bf217e837d13efe2f3ccc3042"

  def install
    bin.install "codacy-cli.sh" => "codacy-cli"
  end
end
