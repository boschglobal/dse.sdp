// @ts-check
const js = require('@eslint/js');
const tseslint = require('typescript-eslint');

module.exports = tseslint.config(
    { ignores: ['**/node_modules/**', 'client/out/**', 'server/out/**', 'out/**', 'bin/**'] },
    js.configs.recommended,
    ...tseslint.configs.recommended,
    {
        rules: {
            'semi': ['error', 'always'],
            '@typescript-eslint/no-unused-vars': 'off',
            '@typescript-eslint/no-explicit-any': 'off',
            '@typescript-eslint/explicit-module-boundary-types': 'off',
            '@typescript-eslint/no-non-null-assertion': 'off',
            // Newly recommended in ESLint 9/10; existing code predates them.
            'prefer-const': 'warn',
            'no-useless-escape': 'warn',
            'no-useless-assignment': 'warn',
            '@typescript-eslint/no-empty-object-type': 'warn',
            '@typescript-eslint/no-require-imports': 'warn',
            '@typescript-eslint/no-unused-expressions': 'warn',
        },
    },
);
