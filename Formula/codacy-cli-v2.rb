class CodacyCliV2 < Formula
  version "1.0.0-main.377.75d97e9"
  url "https://raw.githubusercontent.com/codacy/codacy-cli-v2/1.0.0-main.377.sha.75d97e9/codacy-cli.sh"
  sha256 "1bb82234e74e5385ae6a6e93cb61cbc3356fdffdeb36f907686b88679c0cd82c"

  def install
    bin.install "codacy-cli.sh" => "codacy-cli"
  end
end
