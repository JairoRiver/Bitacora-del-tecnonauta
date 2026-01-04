import type { ParagraphType } from "../data/helpers/postsData"

export const TAILWIND_PARAGRAPH_MAP: Record<ParagraphType, string> = {
    'size-lg': 'text-lg',
    'bold': 'font-bold',
    'italic': 'italic',
    'underline': 'underline',
    'color-1': 'text-dark-principal dark:text-light-principal',
    'color-2': 'text-dark-secundary dark:text-light-secundary',
    'color-3': 'text-dark-third dark:text-light-third',
};