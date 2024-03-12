package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	InitTemp "pokemon/temp"
	"strings"
	"time"
)

var AllPokemonTypes []string

type Pokemon struct {
	ID              int             `json:"id"`
	Name            string          `json:"name"`
	URL             string          `json:"url"`
	Type            []string        `json:"types"`
	Image           string          `json:"image"`
	Abilities       []Ability       `json:"abilities"` // Nouveau champ pour les capacités
	DamageRelations DamageRelations `json:"damage_relations"`
}

type Ability struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ApiResponse struct {
	Count    int       `json:"count"`
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
	Results  []Pokemon `json:"results"`
}

type DamageRelations struct {
	DoubleDamageFrom []TypeRelation `json:"double_damage_from"`
	DoubleDamageTo   []TypeRelation `json:"double_damage_to"`
	HalfDamageFrom   []TypeRelation `json:"half_damage_from"`
	HalfDamageTo     []TypeRelation `json:"half_damage_to"`
	NoDamageFrom     []TypeRelation `json:"no_damage_from"`
	NoDamageTo       []TypeRelation `json:"no_damage_to"`
}

type TypeRelation struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func ToLower(str string) string {
	return strings.ToLower(str)
}

func init() {
	// Initialisation de AllPokemonTypes au démarrage de l'application.
	FetchPokemonTypes()
}

func FetchPokemonTypes() {
	// Effectuez la requête à l'API pour récupérer les types de Pokémon.
	resp, err := http.Get("https://pokeapi.co/api/v2/type/")
	if err != nil {
		log.Fatalf("Erreur lors de la récupération des types de Pokémon : %v", err)
	}
	defer resp.Body.Close()

	// Parsez la réponse JSON.
	var data struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatalf("Erreur lors du décodage de la réponse JSON: %v", err)
	}

	// Remplissez AllPokemonTypes avec les noms des types de Pokémon.
	for _, t := range data.Results {
		AllPokemonTypes = append(AllPokemonTypes, t.Name)
	}
}

// Supposons que cette fonction envoie une requête à l'API et récupère 20 Pokémon aléatoires
func GetRandomPokemons() ([]Pokemon, error) {
	rand.Seed(time.Now().UnixNano()) // Initialise le générateur de nombres aléatoires

	var pokemons []Pokemon

	for i := 0; i < 20; i++ {
		// Générer un ID aléatoire pour un Pokémon. Ajustez le max selon le nombre total de Pokémon disponibles.
		pokemonID := rand.Intn(898) + 1
		pokemonURL := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%d", pokemonID)

		// La fonction FetchPokemonDetails est modifiée pour retourner le nom, types, abilities, et l'image du Pokémon.
		id, name, types, abilities, image, err := FetchPokemonDetails(pokemonURL)
		if err != nil {
			log.Printf("Failed to fetch Pokemon details for ID %d: %v", pokemonID, err)
			continue // Continue to the next iteration if an error occurs
		}

		// Crée un nouvel objet Pokémon avec les détails récupérés et l'ajoute à la liste
		pokemon := Pokemon{
			ID:        id,
			Name:      name,
			Type:      types,
			Abilities: abilities, // Assurez-vous d'ajouter les capacités ici
			Image:     image,
		}
		pokemons = append(pokemons, pokemon)
	}

	return pokemons, nil
}

func FetchPokemonsForType(typeName string) ([]Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/type/%s", typeName)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching type %s: %v", typeName, err)
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Pokemon []struct {
			Pokemon struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"pokemon"`
		} `json:"pokemon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var pokemons []Pokemon
	for _, p := range result.Pokemon {
		// Optionnellement, récupérez plus de détails ici avec FetchPokemonDetails
		pokemons = append(pokemons, Pokemon{Name: p.Pokemon.Name})
	}

	return pokemons, nil
}

func FetchPokemonDetails(pokemonURL string) (id int, name string, types []string, abilities []Ability, image string, err error) {
	resp, err := http.Get(pokemonURL)
	if err != nil {
		return 0, "", nil, nil, "", err
	}
	defer resp.Body.Close()

	var detailResp struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Sprites struct {
			Other struct {
				OfficialArtwork struct {
					FrontDefault string `json:"front_default"`
				} `json:"official-artwork"`
			} `json:"other"`
		} `json:"sprites"`
		Types []struct {
			Type struct {
				Name string `json:"name"`
			} `json:"type"`
		} `json:"types"`
		Abilities []struct {
			Ability struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"ability"`
		} `json:"abilities"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&detailResp); err != nil {
		return 0, "", nil, nil, "", err
	}

	id = detailResp.ID
	name = detailResp.Name
	image = detailResp.Sprites.Other.OfficialArtwork.FrontDefault

	for _, t := range detailResp.Types {
		types = append(types, t.Type.Name)
	}

	for _, a := range detailResp.Abilities {
		abilities = append(abilities, Ability{
			Name: a.Ability.Name,
			URL:  a.Ability.URL,
		})
	}

	return id, name, types, abilities, image, nil
}

func Index(w http.ResponseWriter, r *http.Request) {
	pokemons, err := GetRandomPokemons()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des Pokémon", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Nombre de Pokémons récupérés : %d\n", len(pokemons))
	// Passez les Pokémon au template
	InitTemp.Temp.ExecuteTemplate(w, "index", pokemons)
}

func SearchPokemon(w http.ResponseWriter, r *http.Request) {
	// Extrait le terme de recherche de la requête
	searchQuery := r.URL.Query().Get("query")
	searchQuery = ToLower(searchQuery)
	if searchQuery == "" {
		http.Error(w, "Vous devez fournir un terme de recherche", http.StatusBadRequest)
		return
	}

	// Utilisez searchQuery pour faire une requête à l'API et obtenir des données sur le Pokémon
	pokemonURL := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", searchQuery)
	id, name, types, abilities, image, err := FetchPokemonDetails(pokemonURL) // Modifié pour inclure les abilities
	if err != nil {
		log.Printf("Failed to fetch details for %s: %v", searchQuery, err) // Add logging here
		if strings.Contains(err.Error(), "404") {
			InitTemp.Temp.ExecuteTemplate(w, "search", map[string]string{
				"ErrorMessage": "Aucun Pokémon trouvé",
			})
		} else {
			http.Error(w, fmt.Sprintf("Erreur lors de la récupération des détails de Pokémon: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Si un Pokémon est trouvé, continuez comme d'habitude
	pokemon := Pokemon{
		ID:        id,
		Name:      name,
		Type:      types,
		Abilities: abilities, // Assurez-vous d'inclure les abilities ici
		Image:     image,
	}

	log.Printf("Fetched details for %s: %+v", searchQuery, pokemon)
	InitTemp.Temp.ExecuteTemplate(w, "search", pokemon)
}

func PokemonDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extrayez le nom du Pokémon de l'URL
	name := strings.TrimPrefix(r.URL.Path, "/pokemon/")

	// Utilisez `FetchPokemonDetails` pour obtenir les détails du Pokémon
	id, name, types, abilities, image, err := FetchPokemonDetails("https://pokeapi.co/api/v2/pokemon/" + name) // Ajouté abilities dans la récupération
	if err != nil {
		http.Error(w, "Pokémon non trouvé", http.StatusNotFound)
		return
	}
	if err != nil {
		// Gérez l'erreur, peut-être en continuant sans les informations d'évolution
		log.Printf("Erreur lors de la récupération des détails d'évolution: %v", err)
	}

	damageRelations, err := FetchTypeDamageRelations(types[0])
	if err != nil {
		log.Printf("Erreur lors de la récupération des relations de dommages pour le type %s: %v", types[0], err)
		// Vous pouvez soit ignorer cette erreur soit retourner une erreur HTTP
	}

	// Créez une instance de Pokémon avec les détails obtenus
	pokemon := Pokemon{
		ID:              id,
		Name:            name,
		Type:            types,
		Abilities:       abilities, // Incluez les capacités ici
		Image:           image,
		DamageRelations: damageRelations,
	}

	// Passez le Pokémon au template de détail
	InitTemp.Temp.ExecuteTemplate(w, "pokemon", pokemon)
}

func FetchEvolutionDetails(evolutionChainID int) (evolutions []string, err error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/evolution-chain/%d/", evolutionChainID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Chain struct {
			EvolvesTo []struct {
				Species struct {
					Name string `json:"name"`
				} `json:"species"`
				EvolvesTo []struct {
					Species struct {
						Name string `json:"name"`
					} `json:"species"`
				} `json:"evolves_to"`
			} `json:"evolves_to"`
		} `json:"chain"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Ajouter le premier Pokémon (base) à la liste des évolutions
	evolutions = append(evolutions, data.Chain.EvolvesTo[0].Species.Name)

	// Parcourir la chaîne d'évolution et ajouter chaque évolution
	for _, evolution := range data.Chain.EvolvesTo {
		if len(evolution.EvolvesTo) > 0 {
			evolutions = append(evolutions, evolution.EvolvesTo[0].Species.Name)
		}
	}

	return evolutions, nil
}

func FetchTypeDamageRelations(typeName string) (DamageRelations, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/type/%s", typeName)
	resp, err := http.Get(url)
	if err != nil {
		return DamageRelations{}, err
	}
	defer resp.Body.Close()

	var damageRelations DamageRelations
	if err := json.NewDecoder(resp.Body).Decode(&damageRelations); err != nil {
		return DamageRelations{}, err
	}

	return damageRelations, nil
}
