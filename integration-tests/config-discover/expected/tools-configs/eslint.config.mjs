export default [
    {
        rules: {
          "constructor-super": ["error"],
          "for-direction": ["error"],
          "getter-return": ["error", {"allowImplicit": false}],
          "no-async-promise-executor": ["error"],
          "no-useless-backreference": ["error"],
          "no-useless-catch": ["error"],
          "no-useless-escape": ["error"],
          "no-with": ["error"],
          "require-yield": ["error"],
          "use-isnan": ["error", {"enforceForIndexOf": false, "enforceForSwitchCase": true}],
          "valid-typeof": ["error", {"requireStringLiterals": false}],
        }
    }
];