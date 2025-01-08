import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';
import importPlugin from 'eslint-plugin-import';

export default tseslint.config([
    {
        ignores: [
            "**/*.pb.ts",
            "**/*.js",
            "**/*.mjs"
        ],
    },
    {
        extends: [
            eslint.configs.recommended,
            importPlugin.flatConfigs.errors,
            importPlugin.flatConfigs.warnings,
            tseslint.configs.recommended,
            importPlugin.flatConfigs.typescript,
        ],
        rules: {
            "import/default": 0,
            "import/no-named-as-default-member": 0,
            "import/named": 2,
            "import/order": [
                2,
                {
                    alphabetize: {
                        order: "asc",
                        caseInsensitive: true,
                    },
                }
            ],
            "@typescript-eslint/explicit-module-boundary-types": 0,
            "@typescript-eslint/no-explicit-any": 0,
            "@typescript-eslint/ban-ts-comment": 0,
            "import/no-named-as-default": 0,
            "@typescript-eslint/switch-exhaustiveness-check": ["error",
                {
                    "considerDefaultExhaustiveForUnions": true
                }],
            "import/no-unresolved": 0,
        },
        languageOptions: {
            ecmaVersion: 5,
            sourceType: "script",
            parserOptions: {
                project: "./tsconfig.json",
            },
        },
    }
]);
