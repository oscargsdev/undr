package com.undr.demo.bootstrap;

import com.undr.demo.domain.*;
import com.undr.demo.domain.enums.ProjectStatusEnum;
import com.undr.demo.dto.ProjectCreationDTO;
import com.undr.demo.repository.*;
import com.undr.demo.service.ProjectService;
import jakarta.transaction.Transactional;
import lombok.extern.slf4j.Slf4j;
import org.hibernate.SessionFactory;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import java.time.LocalDate;
import java.time.temporal.ChronoUnit;
import java.util.Arrays;
import java.util.List;
import java.util.Random;
import java.util.concurrent.ThreadLocalRandom;
import java.util.stream.Collectors;

@Slf4j
@Component
public class BootstrapData implements CommandLineRunner {
    private final ProjectService projectService;
    private final ProjectGenreRepository projectGenreRepository;
    private final ProjectLocationRepository projectLocationRepository;
    private final StreamingLinksRepository streamingLinksRepository;
    private final SocialLinksRepository socialLinksRepository;
    private final MusicGenreRepository musicGenreRepository;
    private final MusicGenreDummyData musicGenreDummyData;


    private final SessionFactory sessionFactory;

    public BootstrapData(
            ProjectService projectService,
            ProjectGenreRepository projectGenreRepository,
            ProjectLocationRepository projectLocationRepository,
            StreamingLinksRepository streamingLinksRepository,
            SocialLinksRepository socialLinksRepository,
            MusicGenreRepository musicGenreRepository,
            SessionFactory sessionFactory){
        this.projectService = projectService;

        this.projectGenreRepository = projectGenreRepository;
        this.projectLocationRepository = projectLocationRepository;
        this.streamingLinksRepository = streamingLinksRepository;
        this.socialLinksRepository = socialLinksRepository;
        this.sessionFactory = sessionFactory;
        this.musicGenreRepository = musicGenreRepository;
        musicGenreDummyData = new MusicGenreDummyData(this.musicGenreRepository);
        musicGenreDummyData.insertInDB();
    }
    private static final Random RANDOM = new Random();
    public static <T extends Enum<?>> T getRandomEnum(Class<T> clazz) {
        T[] values = clazz.getEnumConstants();
        return values[RANDOM.nextInt(values.length)];
    }

    public static LocalDate getRandomDateBetween1969AndNow() {
        LocalDate start = LocalDate.of(1969, 1, 1);
        LocalDate end = LocalDate.now();

        long daysBetween = ChronoUnit.DAYS.between(start, end);
        long randomDayOffset = ThreadLocalRandom.current().nextLong(daysBetween + 1);

        return start.plusDays(randomDayOffset);
    }
    @Override
    @Transactional
    public void run(String... args) throws Exception {
        System.out.println("""






                \\n
                IN BOOTSTRAP DATA





                \\n



                """);

        // Projects
        List<String> undergroundBands = Arrays.asList(
                "Grave Moss Ritual",
                "The Velvet Shriek",
                "Concrete Hymns",
                "Rust Lung",
                "Glass Prophet",
                "Swamp Church",
                "Neon Casket",
                "Black Static",
                "Sunless Choir",
                "Burnt Milk",
                "Witch Garage",
                "Tape Spool Dreams",
                "Moth Season",
                "Gutter Saints",
                "Coffin Rodeo",
                "Echo Thief",
                "Strobe Ruin",
                "Ash Vision",
                "Terminal Sleep",
                "Crypt Window",
                "Broken Light Fixture",
                "Gnawed Silence",
                "Hollow Format",
                "Dogtooth Oracle",
                "Vein Prism",
                "Basement Epiphany",
                "Feral Maps",
                "Plastic Psalm",
                "Doomfax",
                "Nocturnal Glitch",
                "Brittle Bone Disco",
                "Shrine of Dirt",
                "The Mold Harvest",
                "Signal Bleed",
                "Velvet Denture",
                "Bleach Mirage",
                "Ghost Clipboard",
                "Yawn Cult",
                "The Stains of Saturn",
                "Drift Coffin",
                "Nail Shower",
                "Wiretongue",
                "Subzero Monk",
                "Crate Burial",
                "Ember Wallet",
                "Dead Channel Youth",
                "The Soft Collapse",
                "Jagged Choir",
                "Obsidian Arcade",
                "Cursed Wallpaper",
                "Midnight Filing Cabinet",
                "False Heaven Broadcast",
                "Casette Wound",
                "The Sun is a Trap",
                "Gloom Tap",
                "Cracked Badge",
                "Posthumous Glow",
                "Lurid Wake",
                "Neon Grudge",
                "Thrift Store Messiah",
                "The Static Knell",
                "Broken Synth Confessional",
                "Sulfur Paradox",
                "Data Rot Saints",
                "Curtain of Moths",
                "Disco Embers",
                "Haunted Payroll",
                "Knuckle Gospel",
                "Crushed Velvet Orphanage",
                "Echo Chamber Monk",
                "Vacuum Ritual",
                "Chipped Halo",
                "The Outdated Index",
                "Detuned Baptism",
                "Synthetic Blight",
                "Mildew Prophet",
                "Erratic Sleep Patterns",
                "The Panic Clergy",
                "Dilapidated Drone",
                "Grime Pulp",
                "Uncanny Workshop",
                "Dialtone Bones",
                "The Forgotten Patch Cable",
                "Binary Chapel",
                "Bleeding Polaroids",
                "The Dissonant Table",
                "Tape Loop Priest",
                "Resin Wounds",
                "Iron Thistle",
                "Cradle of Dust",
                "Frostbitten Debris",
                "Pagan LAN Party",
                "Shriek Index",
                "Napalm Carousel",
                "Velvet Plague",
                "Gargoyle Ambush",
                "Rust Rapture",
                "Cardboard Messiah",
                "Tombwave",
                "Exit Wound Parade",
                "Decayed LAN",
                "Noise Baptism",
                "Funeral Spreadsheet",
                "Ghost Tax Audit"
        );

        List<String> undergroundGenres = Arrays.asList(
                "Funeral Surf",
                "Noirwave",
                "Post-Industrial Gospel",
                "Crust Jazz",
                "Lo-fi Doom",
                "Swamp Psych",
                "Haunt Blues",
                "Drone Folk",
                "Acid Shoegaze",
                "Blackgaze Country",
                "Witchpop",
                "Tape Hiss Ambient",
                "Post-Mothcore",
                "Trashwave",
                "Occult Western",
                "Noise Liturgical",
                "Stoner Gospel",
                "Ashcore",
                "Post-Sludge Dream",
                "Crypt Techno",
                "Gothic Ska",
                "Math Crust",
                "Neo-Twee Metal",
                "Experimental Static",
                "Vapor Doom",
                "Dreamviolence",
                "Psychofolk Dub",
                "Cybergrind Jazz",
                "Hollowwave",
                "Grindgaze",
                "Obscure Pop Terror",
                "Grave Disco",
                "Coffinbeat",
                "Digital Drone Punk",
                "Slacker Doom",
                "Ethereal Noise Rap",
                "Deadwave",
                "Spiritual Slowcore",
                "Post-Synth Ritual",
                "Glitchgaze",
                "Ritual Ska Doom",
                "Haunted Crunk",
                "Subterranean Shoegaze",
                "Drone ‘n’ Bass",
                "Burial Country",
                "Twee Death",
                "Grimegaze",
                "Post-Vapour Trap",
                "Softcore Black Metal",
                "Field Recording Pop",
                "Analog Doom Folk",
                "Post-Dada Noise",
                "Lo-Baptist Drone",
                "Cassettecore",
                "Decomposed Soul",
                "Distorted IDM",
                "Occult Trap",
                "Industrial Samba",
                "Dungeonwave",
                "Ghostbeat",
                "Cybercowboy Doom",
                "Glitched Alt Folk",
                "Staticcore",
                "Avant-Sludge",
                "Cringe Folk",
                "Neo-Basement Rock",
                "Inverted Jazz",
                "Underpass Punk",
                "Voidwave",
                "Noisechant",
                "Micropunk",
                "Proto-Drone",
                "Witchstep",
                "Mid-Fi Techno",
                "Arcadecore",
                "Grungy Noise Gospel",
                "Laptop Shoegaze",
                "Post-Post Punk",
                "Minimalist Gothwave",
                "Distorted Americana",
                "Experimental Garagecore",
                "Ambient Scuzz",
                "Tape Death",
                "Crushed Soul Trap",
                "Glitchwave Revival",
                "Mycelium Folk",
                "Mildewcore",
                "Bitcrush Ambient",
                "Analog Scream Jazz",
                "Horror Folkstep",
                "Neo-Crunch",
                "Doom Punk Gospel",
                "Pixelgrind",
                "Melancholy Ska Doom",
                "Post-Utility Rock",
                "Field Doom",
                "Shoegaze Trapcore",
                "Cursed Chillhop",
                "Noise Folk Funk",
                "Deadbeat Grunge",
                "Minimal Doom Jazz",
                "Hypnagogic Metal",
                "Funeraltrap"
        );

        List<List<String>> undergroundSubgenres = Arrays.asList(
                Arrays.asList("Surf Doom", "Beach Crust", "Tropical Black Metal"),
                Arrays.asList("Neo-Noir Pop", "Shadowwave", "Cinematic Synthpunk"),
                Arrays.asList("Industrial Gospelcore", "Post-Church Rock", "Grime Hymnals"),
                Arrays.asList("Jazzcore", "Crust-Fusion", "Sludge Improv"),
                Arrays.asList("Doomwave", "Lo-fi Funeral Punk"),
                Arrays.asList("Bayou Shoegaze", "Bog Rock", "Marshcore"),
                Arrays.asList("Ghost Blues", "Seance Funk", "Spectral Roots"),
                Arrays.asList("Folktronics", "Dronegrass", "Slowcore Americana"),
                Arrays.asList("Acid Shoepop", "Hazy Glitchgaze", "Dream Acid"),
                Arrays.asList("Countrygaze", "Alt-Grass", "Cinematic Westerncore"),
                Arrays.asList("Pagan Pop", "Witchtrap", "Occult Lounge"),
                Arrays.asList("Ambient Lo-Fi Tape", "Spoolcore", "Analog Whisperwave"),
                Arrays.asList("Mothwave", "Fuzzed Psychfolk", "Gossamercore"),
                Arrays.asList("Trashgaze", "Artpunk Sludge", "Noiseglam"),
                Arrays.asList("Doom Western", "Occult Country", "Grave Rodeo"),
                Arrays.asList("Sacred Feedback", "Liturgical Drone", "Holy Static"),
                Arrays.asList("Desert Stoner Gospel", "Baptized Fuzz", "Smoke Sermons"),
                Arrays.asList("Ash Ambient", "Burncore", "Post-Cremation Electronica"),
                Arrays.asList("Dreamsludge", "Shoethrob", "Melted Gloom"),
                Arrays.asList("Dark Rave", "Crypt EDM", "Grave Synthstep"),
                Arrays.asList("Skathic", "Goth Brass", "Two-Tone Doom"),
                Arrays.asList("Crust Math Rock", "Fractured Punk", "Hard Equationcore"),
                Arrays.asList("Indie Doom Twee", "Softblack", "Glitch Twee"),
                Arrays.asList("Static Noise", "Experimental Silence", "Post-Digital Drone"),
                Arrays.asList("Slowdoom Vapor", "Fume Pop", "Dronestep"),
                Arrays.asList("Dreamviolence", "Blisscore", "Sadcore Thrash"),
                Arrays.asList("Doom Dub Folk", "Post-Fusion Folk", "Echo Rituals"),
                Arrays.asList("CyberJazzgrind", "Mathnoise", "Digital Throttlecore"),
                Arrays.asList("Shoegloom", "Hollow Dream", "Gaze Flux"),
                Arrays.asList("Gazegrind", "Blistergaze", "Broken Screamo"),
                Arrays.asList("Noise Pop Horror", "Panic Indie", "Dreadwave"),
                Arrays.asList("Gravewave", "Dark Disco Funk", "Gravedigger Groove"),
                Arrays.asList("Epitaphcore", "Grimpop", "Ghoulstep"),
                Arrays.asList("Dronepunk", "Analog Collapse", "Digital Decay"),
                Arrays.asList("Doom Chill", "Post-Psych Lounge", "Stoner Softcore"),
                Arrays.asList("Noise Rap Psalm", "Trap Lament", "Gospel Trap Gaze"),
                Arrays.asList("Deadcore", "Post-Mortem Electronica", "Drone of Silence"),
                Arrays.asList("Soft Ritual", "Sacredcore", "Ambient Gospel"),
                Arrays.asList("Synth Rites", "Vapor Occult", "Witchtempo"),
                Arrays.asList("Dreamgaze", "Cyber Drone", "Post-Sleep Shoegaze"),
                Arrays.asList("Ritual Ska Fusion", "Skadoom", "Heavy Brass Drones"),
                Arrays.asList("Crunk Horror", "Swamp Trap", "Screamo Bounce"),
                Arrays.asList("Underground Shoepunk", "Metro Fuzz", "Basement Dream"),
                Arrays.asList("DNB Drone", "Rave Decay", "Wubcore"),
                Arrays.asList("Cowboy Doomgrass", "Grave Rodeo Funk", "Alt-Country Noise"),
                Arrays.asList("Death Twee", "Sad Glitter Pop", "Underground Sparkle"),
                Arrays.asList("Grime Pop", "Muckwave", "Sludge Electro"),
                Arrays.asList("Trap Vapourwave", "Decay-Hop", "Echo-Drill"),
                Arrays.asList("Black Metal Easycore", "Gloom Bubble Punk"),
                Arrays.asList("Tape Pop", "Glitch Records", "Cassette Chill"),
                Arrays.asList("Acoustic Drone Doom", "Analog Slowfolk", "Ghost Strings"),
                Arrays.asList("Noise Dada Rock", "Absurdcore", "Fluxwave"),
                Arrays.asList("Drone Hymnal", "Static Preach", "Ghost Chant"),
                Arrays.asList("Lo-Fi Cassette Rock", "Woundpop", "Magnetcore"),
                Arrays.asList("Rot Pop", "Putrid Shoegaze", "Corpse Disco"),
                Arrays.asList("IDM Decay", "Glitch Rot", "Noise Fragmentation"),
                Arrays.asList("Trap Ritual", "Cult-Hop", "Darkwave Flow"),
                Arrays.asList("Samba Noise", "Industrial Carnival", "Percussive Terror"),
                Arrays.asList("Dungeon Trap", "Castlewave", "Sword & Bass"),
                Arrays.asList("Hauntbeat", "Spectral Groove", "Ghosthop"),
                Arrays.asList("Western Doomcore", "Cybercowboy Punk", "Rustwave"),
                Arrays.asList("Folktronica Fade", "Roots Glitch", "Downtempo Americana"),
                Arrays.asList("Analog Glitchcore", "Static Trap", "Noise Compression"),
                Arrays.asList("Sludge Avant", "Jazzcore Experimental", "Weird Rock"),
                Arrays.asList("Acoustic Sad Folk", "Whimperwave", "Coffeeshop Crust"),
                Arrays.asList("Post-Basement Rock", "Drywallcore", "Garage Goth"),
                Arrays.asList("Free Jazz Implosion", "Experimental Swing Doom"),
                Arrays.asList("Alley Punk", "Street Shoegaze", "Concretecore"),
                Arrays.asList("Void Jazz", "Gothstep", "Ambience Collapse"),
                Arrays.asList("Haunted Rituals", "Choral Drones", "Sacred Noise"),
                Arrays.asList("Microwave Punk", "Small Room Hardcore", "Nano Crust"),
                Arrays.asList("Proto-Droneskate", "Noise Folk Skate", "Basement Thrash"),
                Arrays.asList("Witchbass", "Hexbeat", "Specterstep"),
                Arrays.asList("Mid-Fi Cloudrap", "Middle Fidelity Doom", "Budget Trap"),
                Arrays.asList("Arcade Thrash", "Pixel Doom", "Consolecore"),
                Arrays.asList("Sludge Gospel Noise", "Fuzz Hymns", "Stoner Evangelist"),
                Arrays.asList("Shoegaze Laptop Pop", "Bitgaze", "Digital Dreamgloom"),
                Arrays.asList("New Post-Punk", "Post of Post", "Endless Echocore"),
                Arrays.asList("Minimal Gothwave", "Lo-Wave", "Grayscale Synthpop"),
                Arrays.asList("Neo-Americana", "Rusty Dream Pop", "Static Folk"),
                Arrays.asList("Garagecore Revival", "Twanggrind", "DIY Doomgrass"),
                Arrays.asList("Scuzzwave", "Lo-Fi Mud Rock", "Garage Haze"),
                Arrays.asList("C-Drip", "Cassette Drone", "Memewave"),
                Arrays.asList("Trap Liturgy", "Sacred Bass", "Masshop"),
                Arrays.asList("Glitchwave Nostalgia", "Vapor Drone Rock", "Beta Pop"),
                Arrays.asList("Folk Mycelia", "Sporesynth", "Mosswave"),
                Arrays.asList("Rotcore", "Mildew Metal", "Moistwave"),
                Arrays.asList("Crushed Bit Folk", "8-Bit Lament", "Antique Pop"),
                Arrays.asList("Scream Jazz", "Distorted Lounge", "Feedback Bebop"),
                Arrays.asList("Folkstep Terror", "Campfire Dub", "Cryptograss"),
                Arrays.asList("Neo Grunge Funk", "Crunchwave", "Doom Pop Punk"),
                Arrays.asList("Apocalypse Gospel", "Trap Sermon", "Sludge Liturgy"),
                Arrays.asList("Glitchgrind", "Noisebit", "Techlash Punk"),
                Arrays.asList("Sad Ska Doom", "Trombone Despair", "Emo Brass"),
                Arrays.asList("Utility Rockcore", "Cubicle Punk", "Slackwave"),
                Arrays.asList("Earthtone Doom", "Naturecore", "Barkstep"),
                Arrays.asList("Trapgaze Fusion", "Lo-fi Crunk Dream", "Soft Bass Sorrow"),
                Arrays.asList("Haunted Lofi Hop", "Spectral Chill", "Dustwave"),
                Arrays.asList("Funk Grief", "Folk Noise Funk", "Disco Dirge"),
                Arrays.asList("Slackgrunge", "Deadpan Rock", "Static Pop"),
                Arrays.asList("Bop Doom", "Jazzy Sludge", "Evil Swing"),
                Arrays.asList("Sleep Metal", "Melancholy Drone", "Napcore"),
                Arrays.asList("Trap Lament", "Cryptflow", "Funeral Hop")
        );

        List<List<String>> bandLocations = Arrays.asList(
                Arrays.asList("Aberdeen, WA", "Portland, OR", "Pacific Northwest"),
                Arrays.asList("Bristol, UK", "Berlin, DE", "Western Europe"),
                Arrays.asList("Detroit, MI", "Cleveland, OH", "Rust Belt"),
                Arrays.asList("New Orleans, LA", "Houston, TX", "Gulf Coast"),
                Arrays.asList("Reykjavík, IS", "Oslo, NO", "Nordic Circle"),
                Arrays.asList("Tallahassee, FL", "Atlanta, GA", "Deep South"),
                Arrays.asList("Providence, RI", "Brooklyn, NY", "East Coast DIY"),
                Arrays.asList("Helsinki, FI", "Tartu, EE", "Baltic Scene"),
                Arrays.asList("Manchester, UK", "Sheffield, UK", "Northern England"),
                Arrays.asList("Wellington, NZ", "Melbourne, AU", "Oceania Underground"),
                Arrays.asList("Salem, MA", "Portland, ME", "New England Noir"),
                Arrays.asList("El Paso, TX", "Santa Fe, NM", "Desert Circuit"),
                Arrays.asList("Ghent, BE", "Rotterdam, NL", "Low Countries"),
                Arrays.asList("Buffalo, NY", "Rochester, NY", "Great Lakes Circuit"),
                Arrays.asList("Newcastle, UK", "Leeds, UK", "Midlands Gloom"),
                Arrays.asList("Baltimore, MD", "Philadelphia, PA", "Mid-Atlantic"),
                Arrays.asList("Cincinnati, OH", "Pittsburgh, PA", "Post-Industrial Belt"),
                Arrays.asList("Tucson, AZ", "San Diego, CA", "Border Noise Scene"),
                Arrays.asList("Quebec City, QC", "Montreal, QC", "Francophone Underground"),
                Arrays.asList("Savannah, GA", "Birmingham, AL", "Southern Gothic"),
                Arrays.asList("Des Moines, IA", "Minneapolis, MN", "Midwest Doom Route"),
                Arrays.asList("Cork, IE", "Dublin, IE", "Irish Fringe"),
                Arrays.asList("Warsaw, PL", "Kraków, PL", "Eastern Bloc Revival"),
                Arrays.asList("Vancouver, BC", "Calgary, AB", "Canadian Noise Belt"),
                Arrays.asList("Marseille, FR", "Lyon, FR", "South of France Underground"),
                Arrays.asList("Athens, GR", "Thessaloniki, GR", "Mediterranean Circuit"),
                Arrays.asList("Memphis, TN", "Nashville, TN", "Southern Distortion"),
                Arrays.asList("Lviv, UA", "Kyiv, UA", "Eastern Underground"),
                Arrays.asList("Las Vegas, NV", "Los Angeles, CA", "West Coast Corridor"),
                Arrays.asList("Bogotá, CO", "Medellín, CO", "Andean Alt Scene"),
                Arrays.asList("Tijuana, MX", "Mexico City, MX", "Baja Noise Net"),
                Arrays.asList("Cape Town, ZA", "Johannesburg, ZA", "South African Avant"),
                Arrays.asList("Ljubljana, SI", "Zagreb, HR", "Balkan DIY"),
                Arrays.asList("Chicago, IL", "Milwaukee, WI", "Great Lakes Noise Loop"),
                Arrays.asList("Seville, ES", "Valencia, ES", "Iberian Backline"),
                Arrays.asList("Tokyo, JP", "Osaka, JP", "Noise Axis"),
                Arrays.asList("Lisbon, PT", "Porto, PT", "Iberian Dreamgaze"),
                Arrays.asList("Anchorage, AK", "Seattle, WA", "Northern Drone Circuit"),
                Arrays.asList("Prague, CZ", "Brno, CZ", "Czech Dissonance Network"),
                Arrays.asList("Lima, PE", "Quito, EC", "Pacific Andes Trail"),
                Arrays.asList("Belfast, NI", "Glasgow, UK", "Northern Fringe Circuit"),
                Arrays.asList("Sofia, BG", "Bucharest, RO", "Balkan Bass Route"),
                Arrays.asList("Naples, IT", "Florence, IT", "Italian Doom Triad"),
                Arrays.asList("Jakarta, ID", "Bali, ID", "Island Noise Flow"),
                Arrays.asList("Seattle, WA", "Olympia, WA", "Cascadian Underground"),
                Arrays.asList("Barcelona, ES", "Madrid, ES", "Spanish Experimental Core"),
                Arrays.asList("Belgrade, RS", "Skopje, MK", "Yugo Revival Circuit"),
                Arrays.asList("Istanbul, TR", "Ankara, TR", "Trans-Eurasian Pulse"),
                Arrays.asList("Denver, CO", "Salt Lake City, UT", "Mountain Drone Zone"),
                Arrays.asList("Kuala Lumpur, MY", "Bangkok, TH", "Southeast Noise Alliance"),
                Arrays.asList("Riga, LV", "Vilnius, LT", "Baltic Static Route"),
                Arrays.asList("Columbus, OH", "Ann Arbor, MI", "Midwestern Tape Swap"),
                Arrays.asList("Doha, QA", "Dubai, AE", "Arabian Industrial Fringe"),
                Arrays.asList("Ulaanbaatar, MN", "Almaty, KZ", "Central Asian Rust"),
                Arrays.asList("San Juan, PR", "Havana, CU", "Caribbean Gloomwave"),
                Arrays.asList("Osaka, JP", "Nagoya, JP", "Japanese Feedback Circuit"),
                Arrays.asList("Hobart, AU", "Adelaide, AU", "Tasmanian Lo-Fi Axis"),
                Arrays.asList("Reykjavík, IS", "Akureyri, IS", "Icelandic Isolation Route"),
                Arrays.asList("Zurich, CH", "Geneva, CH", "Alpine Drone Cartel"),
                Arrays.asList("Manila, PH", "Cebu City, PH", "Island Echo Chamber"),
                Arrays.asList("Utrecht, NL", "Amsterdam, NL", "Dutch Noise Web"),
                Arrays.asList("Sarajevo, BA", "Mostar, BA", "Bosnian Post-Everything"),
                Arrays.asList("Tel Aviv, IL", "Haifa, IL", "Levantine Post-Rock"),
                Arrays.asList("Antwerp, BE", "Brussels, BE", "Belgian Reverb Net"),
                Arrays.asList("St. Petersburg, RU", "Moscow, RU", "Russian Haze Highway"),
                Arrays.asList("Buenos Aires, AR", "Montevideo, UY", "Rio de la Plata Drone"),
                Arrays.asList("Bratislava, SK", "Košice, SK", "Slovak Screamo Path"),
                Arrays.asList("Canberra, AU", "Brisbane, AU", "East Coast Loops"),
                Arrays.asList("Santiago, CL", "Valparaíso, CL", "Andean Shoegaze Arc"),
                Arrays.asList("Berlin, DE", "Hamburg, DE", "German Echofront"),
                Arrays.asList("Accra, GH", "Lagos, NG", "West African Noise Net"),
                Arrays.asList("Tunis, TN", "Algiers, DZ", "North African Drone"),
                Arrays.asList("Athens, OH", "Bloomington, IN", "College Town Static"),
                Arrays.asList("Yokohama, JP", "Fukuoka, JP", "Japanese Lo-Fi Ring"),
                Arrays.asList("Luanda, AO", "Maputo, MZ", "Southern African Static"),
                Arrays.asList("Doha, QA", "Amman, JO", "Middle East Mirage Loop"),
                Arrays.asList("Paris, FR", "Nice, FR", "Southern Gaze Union"),
                Arrays.asList("Omaha, NE", "Kansas City, MO", "Plainscore Circuit"),
                Arrays.asList("Dakar, SN", "Bamako, ML", "Saharan Psyche Route"),
                Arrays.asList("New Delhi, IN", "Pune, IN", "Indian Experimental Corridor"),
                Arrays.asList("Minsk, BY", "Vilnius, LT", "Post-Soviet Feedback"),
                Arrays.asList("Innsbruck, AT", "Vienna, AT", "Austrian Noise Trail"),
                Arrays.asList("Jerusalem, IL", "Ramallah, PS", "Holy Land Sound Warps"),
                Arrays.asList("Stavanger, NO", "Trondheim, NO", "Norwegian Fjord Scene"),
                Arrays.asList("Guadalajara, MX", "Oaxaca, MX", "Mexican Drone Valley"),
                Arrays.asList("Richmond, VA", "Chapel Hill, NC", "Southeastern Alt Axis"),
                Arrays.asList("Anchorage, AK", "Fairbanks, AK", "Alaskan Isolationist Route"),
                Arrays.asList("Yellowknife, NT", "Whitehorse, YT", "Arctic Ambient Front"),
                Arrays.asList("Brno, CZ", "Olomouc, CZ", "Moravian Feedback Alliance"),
                Arrays.asList("Leipzig, DE", "Dresden, DE", "Eastern German Gloomline"),
                Arrays.asList("Gothenburg, SE", "Malmö, SE", "Scandinavian Wavefront"),
                Arrays.asList("Medford, OR", "Eureka, CA", "Northern Californian Crust"),
                Arrays.asList("Halifax, NS", "St. John's, NL", "Atlantic Drone Bridge"),
                Arrays.asList("Juneau, AK", "Sitka, AK", "Rainforest Static Coast")
        );

        List<List<String>> bandStreamingLinks = undergroundBands.stream()
                .map(bandName -> {
                    String slug = bandName.toLowerCase()
                            .replaceAll("[^a-z0-9 ]", "")  // remove special characters
                            .replaceAll("\\s+", "-");     // replace spaces with hyphens

                    return Arrays.asList(
                            "spotify.com/" + slug,
                            "tidal.com/" + slug,
                            "applemusic.com/" + slug,
                            "bandcamp.com/" + slug,
                            "soundcloud.com/" + slug
                    );
                })
                .collect(Collectors.toList());



        List<List<String>> bandSocialLinks = undergroundBands.stream()
                .map(bandName -> {
                    String slug = bandName.toLowerCase()
                            .replaceAll("[^a-z0-9 ]", "")  // remove special characters
                            .replaceAll("\\s+", "-");     // replace spaces with hyphens

                    return Arrays.asList(
                            "instagram.com/" + slug,
                            "tiktok.com/" + slug,
                            "youtube.com/" + slug,
                            "facebook.com/" + slug
                    );
                })
                .collect(Collectors.toList());

        Random random = new Random();

//        Project project = new Project();
//        project.setProjectName("undersleep");
//        project.setProjectFoundation(LocalDate.now());
//        project.setStatus(ProjectStatusEnum.ACTIVE);
//        this.projectRepository.save(project);
//
//        Project project2 = new Project();
//        project2.setProjectName("caspio");
//        project2.setProjectFoundation(LocalDate.now());
//        project2.setStatus(ProjectStatusEnum.IN_LIMBO);
//        this.projectRepository.save(project2);
//
//        Project project3 = new Project();
//        project3.setProjectName("epifani");
//        project3.setProjectFoundation(LocalDate.now());
//        project3.setStatus(ProjectStatusEnum.IN_LIMBO);
//        this.projectRepository.save(project3);

        int maxVal = Math.min(undergroundBands.size(), undergroundGenres.size());
        maxVal = Math.min(maxVal, undergroundSubgenres.size());
        maxVal = Math.min(maxVal, bandLocations.size());

        for (int i = 0; i < maxVal; i++){
            ProjectCreationDTO project4 = new ProjectCreationDTO(undergroundBands.get(i), getRandomDateBetween1969AndNow(), getRandomEnum(ProjectStatusEnum.class));
            this.projectService.createProject(project4);

//            ProjectGenre genre4 = new ProjectGenre(random.nextLong(1, 400), random.nextLong(1, 400), random.nextLong(1, 400), random.nextLong(1, 400));
//            genre4.setProject(project4);
//            this.projectGenreRepository.save(genre4);
//
//            ProjectLocation projectLocation4 = new ProjectLocation(
//                    bandLocations.get(i).get(0),
//                    bandLocations.get(i).get(1),
//                    bandLocations.get(i).get(2));
//            projectLocation4.setProject(project4);
//            this.projectLocationRepository.save(projectLocation4);
//
//
//            StreamingLinks streamingLinks4 = new StreamingLinks(
//                    bandStreamingLinks.get(i).get(0),
//                    bandStreamingLinks.get(i).get(1),
//                    bandStreamingLinks.get(i).get(2),
//                    bandStreamingLinks.get(i).get(3),
//                    bandStreamingLinks.get(i).get(4));
//            streamingLinks4.setProject(project4);
//            this.streamingLinksRepository.save(streamingLinks4);
//
//            SocialLinks socialLinks4 = new SocialLinks(
//                    bandSocialLinks.get(i).get(0),
//                    bandSocialLinks.get(i).get(1),
//                    bandSocialLinks.get(i).get(2),
//                    bandSocialLinks.get(i).get(3)
//            );
//            socialLinks4.setProject(project4);
//            this.socialLinksRepository.save(socialLinks4);
        }



//        ProjectGenre genre = new ProjectGenre("shoegaze", "alt", "dreampop");
//        genre.setProject(project);
//        this.projectGenreRepository.save(genre);
//
//        ProjectGenre genre2 = new ProjectGenre("psych rock", "alt");
//        genre2.setProject(project2);
//        this.projectGenreRepository.save(genre2);
//
//        ProjectGenre genre2 = new ProjectGenre("math rock", "midwest emo", "progresiva");
//        genre2.setProject(project2);
//        this.projectGenreRepository.save(genre2);
//

//        ProjectLocation projectLocation = new ProjectLocation("Ocotlan", "Ocotlan", "AMG");
//        projectLocation.setProject(project);
//        this.projectLocationRepository.save(projectLocation);
//
//        ProjectLocation projectLocation2 = new ProjectLocation("Gdl", "Gdl", "AMG");
//        projectLocation2.setProject(project2);
//        this.projectLocationRepository.save(projectLocation2);
//
//        StreamingLinks streamingLinks = new StreamingLinks("spotify.com/undersleep",
//                "tidal.com/undersleep",
//                "applemusic.com/undersleep",
//                "bandcamp.com/undersleep",
//                "soundcloud.com/undersleep");
//
//        streamingLinks.setProject(project);
//        this.streamingLinksRepository.save(streamingLinks);
//
//
//        SocialLinks socialLinks = new SocialLinks(
//                "instagram.com/undrsleep",
//                "tiktok.com/undersleep",
//                "youtube.com/undersleep",
//                "facebook.com/undersleep"
//                );
//        socialLinks.setProject(project);
//        this.socialLinksRepository.save(socialLinks);
    }
}
