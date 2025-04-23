class CodacyCliV2 < Formula
  version "1.0.0-main.248.b0b491e"
  url "https://raw.githubusercontent.com/codacy/codacy-cli-v2/1.0.0-main.248.b0b491e/codacy-cli.sh"
  sha256 "6e1ec511be3c77d49b05b857e0258fc5251507fdf3499b01cd9f2c22c429b1a9"

  def install
    bin.install "codacy-cli.sh" => "codacy-cli"
  end
end
