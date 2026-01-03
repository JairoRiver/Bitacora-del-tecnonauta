import postsRaw from '../posts.json';

type ContentType = 'p' | 'c' | 'i' | 't';

interface ContentBlock {
    type: ContentType;
    value: string;
    conf: string[];
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