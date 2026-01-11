import postsRaw from '../posts.json';

/*
  -p is for paragraphs
  -c is for code
  -i is for image
  -t is for table
*/
export type ParagraphType = 'p';
export type CodeType = 'c';
export type ImageType = 'i';
export type TableType = 't';
export type ParagraphStyleType = 'size-lg' | 'bold' | 'italic' | 'underline' | 'color-1' | 'color-2' | 'color-3';
export type CellStyleType = 'size-lg' | 'size-md' | 'size-xs' | 'bold' | 'italic' | 'underline';
export type LanguageType = 'SQL' | 'TypeScript' | 'Go' | 'Python' | 'Rust';
export type CodeStyleType = 'color-1' | 'color-2' | 'color-3';
export type CodeValueType = { language: LanguageType; text: string; };
export type ImageValueType = { src: string, caption: string };
export type ColumnStyle = { header: string | number; style: CellStyleType[]; colored_numbers?: boolean }
export type TableConf = { has_header: boolean; table_styles: CellStyleType[]; column_style: ColumnStyle[]; }
export type TableValueType = (string | number)[][];
type ContentBlock = TextBlock | CodeBlock | ImageBlock | TableBlock;

interface TextBlock {
    type: ParagraphType;
    value: string;
    conf: ParagraphStyleType[];
}

interface CodeBlock {
    type: CodeType;
    value: CodeValueType;
    conf: CodeStyleType;
}

interface ImageBlock {
    type: ImageType;
    value: ImageValueType
}

interface TableBlock {
    type: TableType,
    value: TableValueType;
    conf: TableConf
}

interface BlogPost {
    Title: string;
    Subtitle: string;
    Date: string;
    Categories: string[];
    HeroImage: string;
    Slug: string;
    Content: ContentBlock[];
}

interface CardData {
    title: string;
    description: string;
    date: string;
    categories: string[];
    heroImage: string;
    slug: string
}

interface BlogPostList {
    Title: string;
    Slug: string;
}

const postsData = (postsRaw as unknown) as BlogPost[];

/**
 * Obtiene una lista de categorías únicas a partir de un array de posts.
 */
export function getCategories(posts: BlogPost[] = postsData): string[] {
    const allCategories = posts.flatMap((post) => post.Categories);

    const uniqueCategories = [...new Set(allCategories)];

    return uniqueCategories.sort((a, b) => a.localeCompare(b));
}

export function getCardInfo(posts: BlogPost[] = postsData): CardData[] {
    return posts.map((post) => {
        // Buscamos el primer bloque de tipo "p" para la descripción
        const firstParagraph = post.Content.find(item => item.type === 'p');

        // Si no hay párrafo, usamos el Subtitle
        const descriptionText = firstParagraph ? firstParagraph.value : post.Subtitle;

        return {
            title: post.Title,
            description: descriptionText,
            date: post.Date,
            categories: post.Categories,
            heroImage: post.HeroImage,
            slug: post.Slug
        };
    });
}

export function parseRouteName(name: string): string {
    return name.toLowerCase().replace(/\s+/g, "-")
}

export function getPostsList(posts: BlogPost[] = postsData): BlogPostList[] {
    return (
        posts.map((post) => {
            return { Title: post.Title, Slug: post.Slug }
        })
    )
}

export function getBlogPost(slug: string, posts: BlogPost[] = postsData): BlogPost {
    const blogPost = posts.find(post => post.Slug === slug) as BlogPost
    return blogPost
}