import type { LanguageType, ParagraphStyleType, CodeStyleType } from "../data/helpers/postsData"
import type { ComponentProps } from 'astro/types';
import { Code } from "astro/components";
type CodeProps = ComponentProps<typeof Code>;
type AstroCodeLanguage = CodeProps['lang'];
type AstroCodeTheme = CodeProps['theme'];

export const TAILWIND_PARAGRAPH_MAP: Record<ParagraphStyleType, string> = {
    'size-lg': 'text-lg',
    'bold': 'font-bold',
    'italic': 'italic',
    'underline': 'underline',
    'color-1': 'text-dark-text-primary dark:text-light-text-primary',
    'color-2': 'text-dark-text-secundary dark:text-light-text-secundary',
    'color-3': 'text-dark-text-third dark:text-light-text-third',
};

export const LANGUAGE_STYLE_MAP: Record<LanguageType, AstroCodeLanguage> = {
    'SQL': 'sql',
    'TypeScript': 'typescript',
    'Rust': 'rust',
    'Python': 'python',
    'Go': 'go'
};

/* the themes list are: https://shiki.style/themes */
export const CODE_STYLE_MAP: Record<CodeStyleType, AstroCodeTheme> = {
    'color-1': 'nord',
    'color-2': 'monokai',
    'color-3': 'light-plus'
}