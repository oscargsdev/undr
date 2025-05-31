package com.undr.demo.bootstrap;

import com.undr.demo.domain.enums.ProjectStatusEnum;
import com.undr.demo.dto.ProjectCreationDTO;
import com.undr.demo.service.MusicGenreService;
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
    private final MusicGenreDummyData musicGenreDummyData;
    private final MusicGenreService musicGenreService;


    private final SessionFactory sessionFactory;

    public BootstrapData(ProjectService projectService, MusicGenreService musicGenreService, SessionFactory sessionFactory) {
        this.projectService = projectService;
        this.musicGenreService = musicGenreService;
        this.sessionFactory = sessionFactory;

        musicGenreDummyData = new MusicGenreDummyData(this.musicGenreService);
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
        List<String> undergroundBands = Arrays.asList("Grave Moss Ritual", "The Velvet Shriek", "Concrete Hymns", "Rust Lung", "Glass Prophet", "Swamp Church", "Neon Casket", "Black Static", "Sunless Choir", "Burnt Milk", "Witch Garage", "Tape Spool Dreams", "Moth Season", "Gutter Saints", "Coffin Rodeo", "Echo Thief", "Strobe Ruin", "Ash Vision", "Terminal Sleep", "Crypt Window", "Broken Light Fixture", "Gnawed Silence", "Hollow Format", "Dogtooth Oracle", "Vein Prism", "Basement Epiphany", "Feral Maps", "Plastic Psalm", "Doomfax", "Nocturnal Glitch", "Brittle Bone Disco", "Shrine of Dirt", "The Mold Harvest", "Signal Bleed", "Velvet Denture", "Bleach Mirage", "Ghost Clipboard", "Yawn Cult", "The Stains of Saturn", "Drift Coffin", "Nail Shower", "Wiretongue", "Subzero Monk", "Crate Burial", "Ember Wallet", "Dead Channel Youth", "The Soft Collapse", "Jagged Choir", "Obsidian Arcade", "Cursed Wallpaper", "Midnight Filing Cabinet", "False Heaven Broadcast", "Casette Wound", "The Sun is a Trap", "Gloom Tap", "Cracked Badge", "Posthumous Glow", "Lurid Wake", "Neon Grudge", "Thrift Store Messiah", "The Static Knell", "Broken Synth Confessional", "Sulfur Paradox", "Data Rot Saints", "Curtain of Moths", "Disco Embers", "Haunted Payroll", "Knuckle Gospel", "Crushed Velvet Orphanage", "Echo Chamber Monk", "Vacuum Ritual", "Chipped Halo", "The Outdated Index", "Detuned Baptism", "Synthetic Blight", "Mildew Prophet", "Erratic Sleep Patterns", "The Panic Clergy", "Dilapidated Drone", "Grime Pulp", "Uncanny Workshop", "Dialtone Bones", "The Forgotten Patch Cable", "Binary Chapel", "Bleeding Polaroids", "The Dissonant Table", "Tape Loop Priest", "Resin Wounds", "Iron Thistle", "Cradle of Dust", "Frostbitten Debris", "Pagan LAN Party", "Shriek Index", "Napalm Carousel", "Velvet Plague", "Gargoyle Ambush", "Rust Rapture", "Cardboard Messiah", "Tombwave", "Exit Wound Parade", "Decayed LAN", "Noise Baptism", "Funeral Spreadsheet", "Ghost Tax Audit");

        List<List<String>> bandLocations = Arrays.asList(Arrays.asList("Aberdeen, WA", "Portland, OR", "Pacific Northwest"), Arrays.asList("Bristol, UK", "Berlin, DE", "Western Europe"), Arrays.asList("Detroit, MI", "Cleveland, OH", "Rust Belt"), Arrays.asList("New Orleans, LA", "Houston, TX", "Gulf Coast"), Arrays.asList("Reykjavík, IS", "Oslo, NO", "Nordic Circle"), Arrays.asList("Tallahassee, FL", "Atlanta, GA", "Deep South"), Arrays.asList("Providence, RI", "Brooklyn, NY", "East Coast DIY"), Arrays.asList("Helsinki, FI", "Tartu, EE", "Baltic Scene"), Arrays.asList("Manchester, UK", "Sheffield, UK", "Northern England"), Arrays.asList("Wellington, NZ", "Melbourne, AU", "Oceania Underground"), Arrays.asList("Salem, MA", "Portland, ME", "New England Noir"), Arrays.asList("El Paso, TX", "Santa Fe, NM", "Desert Circuit"), Arrays.asList("Ghent, BE", "Rotterdam, NL", "Low Countries"), Arrays.asList("Buffalo, NY", "Rochester, NY", "Great Lakes Circuit"), Arrays.asList("Newcastle, UK", "Leeds, UK", "Midlands Gloom"), Arrays.asList("Baltimore, MD", "Philadelphia, PA", "Mid-Atlantic"), Arrays.asList("Cincinnati, OH", "Pittsburgh, PA", "Post-Industrial Belt"), Arrays.asList("Tucson, AZ", "San Diego, CA", "Border Noise Scene"), Arrays.asList("Quebec City, QC", "Montreal, QC", "Francophone Underground"), Arrays.asList("Savannah, GA", "Birmingham, AL", "Southern Gothic"), Arrays.asList("Des Moines, IA", "Minneapolis, MN", "Midwest Doom Route"), Arrays.asList("Cork, IE", "Dublin, IE", "Irish Fringe"), Arrays.asList("Warsaw, PL", "Kraków, PL", "Eastern Bloc Revival"), Arrays.asList("Vancouver, BC", "Calgary, AB", "Canadian Noise Belt"), Arrays.asList("Marseille, FR", "Lyon, FR", "South of France Underground"), Arrays.asList("Athens, GR", "Thessaloniki, GR", "Mediterranean Circuit"), Arrays.asList("Memphis, TN", "Nashville, TN", "Southern Distortion"), Arrays.asList("Lviv, UA", "Kyiv, UA", "Eastern Underground"), Arrays.asList("Las Vegas, NV", "Los Angeles, CA", "West Coast Corridor"), Arrays.asList("Bogotá, CO", "Medellín, CO", "Andean Alt Scene"), Arrays.asList("Tijuana, MX", "Mexico City, MX", "Baja Noise Net"), Arrays.asList("Cape Town, ZA", "Johannesburg, ZA", "South African Avant"), Arrays.asList("Ljubljana, SI", "Zagreb, HR", "Balkan DIY"), Arrays.asList("Chicago, IL", "Milwaukee, WI", "Great Lakes Noise Loop"), Arrays.asList("Seville, ES", "Valencia, ES", "Iberian Backline"), Arrays.asList("Tokyo, JP", "Osaka, JP", "Noise Axis"), Arrays.asList("Lisbon, PT", "Porto, PT", "Iberian Dreamgaze"), Arrays.asList("Anchorage, AK", "Seattle, WA", "Northern Drone Circuit"), Arrays.asList("Prague, CZ", "Brno, CZ", "Czech Dissonance Network"), Arrays.asList("Lima, PE", "Quito, EC", "Pacific Andes Trail"), Arrays.asList("Belfast, NI", "Glasgow, UK", "Northern Fringe Circuit"), Arrays.asList("Sofia, BG", "Bucharest, RO", "Balkan Bass Route"), Arrays.asList("Naples, IT", "Florence, IT", "Italian Doom Triad"), Arrays.asList("Jakarta, ID", "Bali, ID", "Island Noise Flow"), Arrays.asList("Seattle, WA", "Olympia, WA", "Cascadian Underground"), Arrays.asList("Barcelona, ES", "Madrid, ES", "Spanish Experimental Core"), Arrays.asList("Belgrade, RS", "Skopje, MK", "Yugo Revival Circuit"), Arrays.asList("Istanbul, TR", "Ankara, TR", "Trans-Eurasian Pulse"), Arrays.asList("Denver, CO", "Salt Lake City, UT", "Mountain Drone Zone"), Arrays.asList("Kuala Lumpur, MY", "Bangkok, TH", "Southeast Noise Alliance"), Arrays.asList("Riga, LV", "Vilnius, LT", "Baltic Static Route"), Arrays.asList("Columbus, OH", "Ann Arbor, MI", "Midwestern Tape Swap"), Arrays.asList("Doha, QA", "Dubai, AE", "Arabian Industrial Fringe"), Arrays.asList("Ulaanbaatar, MN", "Almaty, KZ", "Central Asian Rust"), Arrays.asList("San Juan, PR", "Havana, CU", "Caribbean Gloomwave"), Arrays.asList("Osaka, JP", "Nagoya, JP", "Japanese Feedback Circuit"), Arrays.asList("Hobart, AU", "Adelaide, AU", "Tasmanian Lo-Fi Axis"), Arrays.asList("Reykjavík, IS", "Akureyri, IS", "Icelandic Isolation Route"), Arrays.asList("Zurich, CH", "Geneva, CH", "Alpine Drone Cartel"), Arrays.asList("Manila, PH", "Cebu City, PH", "Island Echo Chamber"), Arrays.asList("Utrecht, NL", "Amsterdam, NL", "Dutch Noise Web"), Arrays.asList("Sarajevo, BA", "Mostar, BA", "Bosnian Post-Everything"), Arrays.asList("Tel Aviv, IL", "Haifa, IL", "Levantine Post-Rock"), Arrays.asList("Antwerp, BE", "Brussels, BE", "Belgian Reverb Net"), Arrays.asList("St. Petersburg, RU", "Moscow, RU", "Russian Haze Highway"), Arrays.asList("Buenos Aires, AR", "Montevideo, UY", "Rio de la Plata Drone"), Arrays.asList("Bratislava, SK", "Košice, SK", "Slovak Screamo Path"), Arrays.asList("Canberra, AU", "Brisbane, AU", "East Coast Loops"), Arrays.asList("Santiago, CL", "Valparaíso, CL", "Andean Shoegaze Arc"), Arrays.asList("Berlin, DE", "Hamburg, DE", "German Echofront"), Arrays.asList("Accra, GH", "Lagos, NG", "West African Noise Net"), Arrays.asList("Tunis, TN", "Algiers, DZ", "North African Drone"), Arrays.asList("Athens, OH", "Bloomington, IN", "College Town Static"), Arrays.asList("Yokohama, JP", "Fukuoka, JP", "Japanese Lo-Fi Ring"), Arrays.asList("Luanda, AO", "Maputo, MZ", "Southern African Static"), Arrays.asList("Doha, QA", "Amman, JO", "Middle East Mirage Loop"), Arrays.asList("Paris, FR", "Nice, FR", "Southern Gaze Union"), Arrays.asList("Omaha, NE", "Kansas City, MO", "Plainscore Circuit"), Arrays.asList("Dakar, SN", "Bamako, ML", "Saharan Psyche Route"), Arrays.asList("New Delhi, IN", "Pune, IN", "Indian Experimental Corridor"), Arrays.asList("Minsk, BY", "Vilnius, LT", "Post-Soviet Feedback"), Arrays.asList("Innsbruck, AT", "Vienna, AT", "Austrian Noise Trail"), Arrays.asList("Jerusalem, IL", "Ramallah, PS", "Holy Land Sound Warps"), Arrays.asList("Stavanger, NO", "Trondheim, NO", "Norwegian Fjord Scene"), Arrays.asList("Guadalajara, MX", "Oaxaca, MX", "Mexican Drone Valley"), Arrays.asList("Richmond, VA", "Chapel Hill, NC", "Southeastern Alt Axis"), Arrays.asList("Anchorage, AK", "Fairbanks, AK", "Alaskan Isolationist Route"), Arrays.asList("Yellowknife, NT", "Whitehorse, YT", "Arctic Ambient Front"), Arrays.asList("Brno, CZ", "Olomouc, CZ", "Moravian Feedback Alliance"), Arrays.asList("Leipzig, DE", "Dresden, DE", "Eastern German Gloomline"), Arrays.asList("Gothenburg, SE", "Malmö, SE", "Scandinavian Wavefront"), Arrays.asList("Medford, OR", "Eureka, CA", "Northern Californian Crust"), Arrays.asList("Halifax, NS", "St. John's, NL", "Atlantic Drone Bridge"), Arrays.asList("Juneau, AK", "Sitka, AK", "Rainforest Static Coast"));

        List<List<String>> bandStreamingLinks = undergroundBands.stream().map(bandName -> {
            String slug = bandName.toLowerCase().replaceAll("[^a-z0-9 ]", "")  // remove special characters
                    .replaceAll("\\s+", "-");     // replace spaces with hyphens

            return Arrays.asList("spotify.com/" + slug, "tidal.com/" + slug, "applemusic.com/" + slug, "bandcamp.com/" + slug, "soundcloud.com/" + slug);
        }).collect(Collectors.toList());


        List<List<String>> bandSocialLinks = undergroundBands.stream().map(bandName -> {
            String slug = bandName.toLowerCase().replaceAll("[^a-z0-9 ]", "")  // remove special characters
                    .replaceAll("\\s+", "-");     // replace spaces with hyphens

            return Arrays.asList("instagram.com/" + slug, "tiktok.com/" + slug, "youtube.com/" + slug, "facebook.com/" + slug);
        }).collect(Collectors.toList());

        Random random = new Random();

        for (int i = 0; i < undergroundBands.size(); i++) {
            ProjectCreationDTO project4 = new ProjectCreationDTO(undergroundBands.get(i), getRandomDateBetween1969AndNow(), getRandomEnum(ProjectStatusEnum.class));
            this.projectService.createProject(project4);
        }
    }
}
