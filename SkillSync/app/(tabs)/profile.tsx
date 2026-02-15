import { StyleSheet, Text, View, Alert, TouchableOpacity, Platform, ScrollView } from 'react-native';
import { Colors } from '@/constants/theme';
import { useColorScheme } from '@/hooks/use-color-scheme';
import { api } from '@/services/api';
import * as DocumentPicker from 'expo-document-picker';
import { useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { useRouter } from 'expo-router';
import { IconSymbol } from '@/components/ui/icon-symbol';
import { JobCard, JobMatch } from '@/components/JobCard';

export default function ProfileScreen() {
    const colorScheme = useColorScheme();
    const theme = Colors[colorScheme ?? 'light'];

    const [uploading, setUploading] = useState(false);
    const [resumeUploaded, setResumeUploaded] = useState(false);
    const [extractedSkills, setExtractedSkills] = useState<string[]>([]);
    const [matches, setMatches] = useState<JobMatch[]>([]);

    const { signOut, user } = useAuth();
    const router = useRouter();

    const handlePickDocument = async () => {
        try {
            const result = await DocumentPicker.getDocumentAsync({
                type: ['application/pdf', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'],
                copyToCacheDirectory: true,
            });

            if (result.canceled) {
                return;
            }

            const file = result.assets[0];
            setUploading(true);
            setExtractedSkills([]);
            setMatches([]);

            try {
                const response = await api.uploadResume(file.uri, file.name, file.mimeType || 'application/pdf');
                setResumeUploaded(true);

                if (response.extracted_skills) {
                    setExtractedSkills(response.extracted_skills);
                }

                if (response.matches) {
                    setMatches(response.matches);
                }

                if (Platform.OS === 'web') {
                    window.alert('Resume uploaded successfully!');
                } else {
                    Alert.alert('Success', 'Resume uploaded! Scroll down to see your skills and matches.');
                }
            } catch (error: any) {
                if (Platform.OS === 'web') {
                    window.alert(error.message || 'Upload failed');
                } else {
                    Alert.alert('Upload Failed', error.message);
                }
            } finally {
                setUploading(false);
            }
        } catch (err) {
            console.log('Unknown error: ', err);
        }
    };

    const handleSignOut = async () => {
        const doSignOut = async () => {
            await signOut();
            router.replace('/auth/sign-in');
        };

        if (Platform.OS === 'web') {
            if (window.confirm('Are you sure you want to sign out?')) {
                await doSignOut();
            }
        } else {
            Alert.alert(
                'Sign Out',
                'Are you sure you want to sign out?',
                [
                    { text: 'Cancel', style: 'cancel' },
                    { text: 'Sign Out', style: 'destructive', onPress: doSignOut },
                ]
            );
        }
    };

    return (
        <ScrollView
            style={[styles.container, { backgroundColor: theme.background }]}
            contentContainerStyle={styles.content}
        >
            {/* User Info */}
            <View style={[styles.avatarCircle, { backgroundColor: theme.tint + '20' }]}>
                <IconSymbol size={40} name="person.fill" color={theme.tint} />
            </View>

            <Text style={[styles.email, { color: theme.text }]}>{user?.email ?? 'Not signed in'}</Text>

            <View style={styles.separator} />

            {/* Resume Section */}
            <View style={styles.section}>
                <Text style={[styles.sectionTitle, { color: theme.text }]}>Resume</Text>

                {resumeUploaded && (
                    <View style={[styles.statusBanner, { backgroundColor: '#22c55e15', borderColor: '#22c55e40' }]}>
                        <IconSymbol size={18} name="checkmark.circle.fill" color="#22c55e" />
                        <Text style={[styles.statusText, { color: '#22c55e' }]}>Resume uploaded</Text>
                    </View>
                )}

                <TouchableOpacity
                    style={[styles.button, { backgroundColor: theme.tint }]}
                    onPress={handlePickDocument}
                    disabled={uploading}
                    activeOpacity={0.7}
                >
                    <IconSymbol size={20} name="doc.text.fill" color="#fff" />
                    <Text style={styles.buttonText}>
                        {uploading ? 'Uploading...' : resumeUploaded ? 'Upload New Resume' : 'Upload Resume'}
                    </Text>
                </TouchableOpacity>

                <Text style={[styles.hint, { color: theme.icon }]}>
                    Supported: PDF, DOCX (max 10MB)
                </Text>
            </View>

            {/* Extracted Skills Section */}
            {extractedSkills.length > 0 && (
                <>
                    <View style={styles.separator} />
                    <View style={styles.section}>
                        <Text style={[styles.sectionTitle, { color: theme.text }]}>Extracted Skills</Text>
                        <View style={styles.skillTags}>
                            {extractedSkills.map((skill, idx) => (
                                <View key={idx} style={[styles.skillTag, { backgroundColor: theme.tint + '15', borderColor: theme.tint + '30' }]}>
                                    <Text style={[styles.skillTagText, { color: theme.tint }]}>{skill}</Text>
                                </View>
                            ))}
                        </View>
                    </View>
                </>
            )}

            {/* Matches Section */}
            {matches.length > 0 && (
                <>
                    <View style={styles.separator} />
                    <View style={styles.section}>
                        <Text style={[styles.sectionTitle, { color: theme.text }]}>Best Matches</Text>
                        {matches.map((job) => (
                            <JobCard key={job.job_id} item={job} />
                        ))}
                    </View>
                </>
            )}

            <View style={styles.separator} />

            {/* Account Section */}
            <View style={styles.section}>
                <Text style={[styles.sectionTitle, { color: theme.text }]}>Account</Text>

                <TouchableOpacity
                    style={[styles.button, styles.signOutButton]}
                    onPress={handleSignOut}
                    activeOpacity={0.7}
                >
                    <IconSymbol size={20} name="rectangle.portrait.and.arrow.right" color="#fff" />
                    <Text style={styles.buttonText}>Sign Out</Text>
                </TouchableOpacity>
            </View>
        </ScrollView>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
    },
    content: {
        alignItems: 'center',
        paddingTop: 60,
        paddingHorizontal: 20,
        paddingBottom: 40,
    },
    avatarCircle: {
        width: 80,
        height: 80,
        borderRadius: 40,
        alignItems: 'center',
        justifyContent: 'center',
        marginBottom: 12,
    },
    email: {
        fontSize: 18,
        fontWeight: '600',
    },
    separator: {
        marginVertical: 24,
        height: 1,
        width: '100%',
        backgroundColor: '#e0e0e0',
    },
    section: {
        width: '100%',
    },
    sectionTitle: {
        fontSize: 18,
        fontWeight: 'bold',
        marginBottom: 12,
    },
    statusBanner: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 8,
        padding: 12,
        borderRadius: 10,
        borderWidth: 1,
        marginBottom: 12,
    },
    statusText: {
        fontSize: 14,
        fontWeight: '600',
    },
    button: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 8,
        padding: 15,
        borderRadius: 12,
        width: '100%',
    },
    buttonText: {
        color: 'white',
        fontSize: 16,
        fontWeight: 'bold',
    },
    signOutButton: {
        backgroundColor: '#ef4444',
    },
    hint: {
        fontSize: 12,
        marginTop: 8,
        textAlign: 'center',
    },
    skillTags: {
        flexDirection: 'row',
        flexWrap: 'wrap',
        gap: 8,
    },
    skillTag: {
        paddingHorizontal: 12,
        paddingVertical: 6,
        borderRadius: 20,
        borderWidth: 1,
    },
    skillTagText: {
        fontSize: 14,
        fontWeight: '500',
    },
});
