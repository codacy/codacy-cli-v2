class CodacyCliV2 < Formula
  version "0.1.0-main.84.d83aad7"
  url "https://raw.githubusercontent.com/codacy/codacy-cli-v2/0.1.0-main.84.d83aad7/codacy-cli.sh"
  sha256 "c295066ba1237eb40439a5c04bda42775d0a5bbf1a270cdf05628d307c40f527"

  def install
    bin.install "codacy-cli.sh" => "codacy-cli"
  end
end
