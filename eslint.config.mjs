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
            "import/named": 2,
            "import/order": [2,
                {
                    alphabetize: {
                        order: "asc",
                        caseInsensitive: true,
                    },
                }
            ],
            "@typescript-eslint/no-explicit-any": 0,
            "@typescript-eslint/ban-ts-comment": 0,
            "import/no-named-as-default": 0,
            "@typescript-eslint/switch-exhaustiveness-check": [2,
                {
                    "considerDefaultExhaustiveForUnions": true
                }],
            "import/no-unresolved": 0,
        },
        languageOptions: {
            ecmaVersion: 2018,
            sourceType: "commonjs",
            parserOptions: {
                project: "./tsconfig.json",
            },
        },
    }
]);
