class CodacyCliV2 < Formula
  version "0.1.0-main.87.6113cd1"
  url "https://raw.githubusercontent.com/codacy/codacy-cli-v2/0.1.0-main.87.6113cd1/codacy-cli.sh"
  sha256 "9cf87b5e81da2e151daf664924502b7d3cdba18260e1126f2be48a5968440585"

  def install
    bin.install "codacy-cli.sh" => "codacy-cli"
  end
end
