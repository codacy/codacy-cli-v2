export default [
    {
        rules: {
          "constructor-super": ["error"],
          "for-direction": ["error"],
          "getter-return": ["error", {"allowImplicit": false}],
          "no-async-promise-executor": ["error"],
          "no-case-declarations": ["error"],
          "no-class-assign": ["error"],
          "no-compare-neg-zero": ["error"],
          "no-cond-assign": ["error", "except-parens"],
          "no-const-assign": ["error"],
          "no-constant-condition": ["error", {"checkLoops": true}],
          "no-control-regex": ["error"],
          "no-debugger": ["error"],
        }
    }
];