import React from 'react';
import { StyleSheet, Text, View, TouchableOpacity, Linking } from 'react-native';
import { Colors } from '@/constants/theme';
import { useColorScheme } from '@/hooks/use-color-scheme';
import { IconSymbol } from '@/components/ui/icon-symbol';

export type JobMatch = {
    job_id: number;
    title: string;
    company: string;
    location: string;
    description: string;
    apply_url: string;
    skill_score: number;
    embed_score: number;
    match_score: number;
    missing_skills: string[] | null;
};

const getScoreColor = (score: number) => {
    if (score >= 0.7) return '#22c55e';
    if (score >= 0.4) return '#f59e0b';
    return '#ef4444';
};

export const JobCard = ({ item }: { item: JobMatch }) => {
    const colorScheme = useColorScheme();
    const theme = Colors[colorScheme ?? 'light'];

    return (
        <View style={[styles.card, { backgroundColor: theme.background, borderColor: theme.icon + '40' }]}>
            <View style={styles.header}>
                <Text style={[styles.jobTitle, { color: theme.text }]}>{item.title}</Text>
                <View style={[styles.badge, { backgroundColor: getScoreColor(item.match_score) }]}>
                    <Text style={styles.badgeText}>{Math.round(item.match_score * 100)}%</Text>
                </View>
            </View>

            <Text style={[styles.company, { color: theme.icon }]}>
                {item.company} • {item.location}
            </Text>

            <Text style={[styles.description, { color: theme.text }]} numberOfLines={3}>
                {item.description}
            </Text>

            {item.missing_skills && item.missing_skills.length > 0 && (
                <View style={styles.skillsContainer}>
                    <Text style={[styles.skillsLabel, { color: theme.icon }]}>Skills to learn:</Text>
                    <View style={styles.skillTags}>
                        {item.missing_skills.slice(0, 5).map((skill, idx) => (
                            <View key={idx} style={[styles.skillTag, { backgroundColor: theme.tint + '20', borderColor: theme.tint + '40' }]}>
                                <Text style={[styles.skillTagText, { color: theme.tint }]}>{skill}</Text>
                            </View>
                        ))}
                        {item.missing_skills.length > 5 && (
                            <Text style={[styles.moreSkills, { color: theme.icon }]}>+{item.missing_skills.length - 5} more</Text>
                        )}
                    </View>
                </View>
            )}

            {item.apply_url ? (
                <TouchableOpacity
                    style={[styles.applyButton, { backgroundColor: theme.tint }]}
                    onPress={() => Linking.openURL(item.apply_url)}
                >
                    <Text style={styles.applyText}>Apply Now</Text>
                    <IconSymbol size={14} name="paperplane.fill" color="#fff" />
                </TouchableOpacity>
            ) : null}
        </View>
    );
};

const styles = StyleSheet.create({
    card: {
        borderWidth: 1,
        borderRadius: 16,
        padding: 16,
        marginBottom: 16,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.08,
        shadowRadius: 8,
        elevation: 3,
    },
    header: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
        marginBottom: 6,
    },
    jobTitle: {
        fontSize: 17,
        fontWeight: '700',
        flex: 1,
        marginRight: 10,
    },
    badge: {
        paddingHorizontal: 10,
        paddingVertical: 4,
        borderRadius: 12,
    },
    badgeText: {
        color: 'white',
        fontWeight: 'bold',
        fontSize: 13,
    },
    company: {
        fontSize: 14,
        marginBottom: 8,
    },
    description: {
        fontSize: 14,
        lineHeight: 20,
        marginBottom: 12,
        opacity: 0.85,
    },
    skillsContainer: {
        marginBottom: 12,
    },
    skillsLabel: {
        fontSize: 12,
        fontWeight: '600',
        marginBottom: 6,
    },
    skillTags: {
        flexDirection: 'row',
        flexWrap: 'wrap',
        gap: 6,
    },
    skillTag: {
        paddingHorizontal: 8,
        paddingVertical: 3,
        borderRadius: 8,
        borderWidth: 1,
    },
    skillTagText: {
        fontSize: 12,
        fontWeight: '500',
    },
    moreSkills: {
        fontSize: 12,
        alignSelf: 'center',
    },
    applyButton: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'center',
        paddingVertical: 10,
        borderRadius: 10,
        gap: 6,
    },
    applyText: {
        color: '#fff',
        fontWeight: '600',
        fontSize: 14,
    },
});
